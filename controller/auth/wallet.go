package auth

import (
	"encoding/hex"
	"errors"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"github.com/yeying-community/router/common"
	"github.com/yeying-community/router/common/config"
	"github.com/yeying-community/router/common/logger"
	"github.com/yeying-community/router/common/random"
	"github.com/yeying-community/router/controller"
	"github.com/yeying-community/router/model"
)

type walletNonceRequest struct {
	Address string `form:"address" json:"address" binding:"required"`
	ChainId string `form:"chain_id" json:"chain_id"`
}

type walletLoginRequest struct {
	Address   string `json:"address"`
	Signature string `json:"signature"`
	Nonce     string `json:"nonce"`
	ChainId   string `json:"chain_id"`
}

// WalletNonce issues a nonce & message to sign
func WalletNonce(c *gin.Context) {
	if !config.WalletLoginEnabled {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "管理员未开启钱包登录",
		})
		return
	}
	var req walletNonceRequest
	if err := c.ShouldBind(&req); err != nil || !common.IsValidEthAddress(req.Address) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "参数错误，缺少 address",
		})
		return
	}

	nonce, message := common.GenerateWalletNonce(req.Address, "Login to "+config.SystemName, req.ChainId)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"nonce":   nonce,
			"message": message,
		},
	})
}

// WalletLogin verifies signature and logs user in
func WalletLogin(c *gin.Context) {
	if !config.WalletLoginEnabled {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "管理员未开启钱包登录",
		})
		return
	}
	var req walletLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "参数错误",
		})
		return
	}

	if err := verifyWalletRequest(req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	addr := strings.ToLower(req.Address)
	// find existing user
	user := model.User{WalletAddress: addr}
	if !model.IsWalletAddressAlreadyTaken(addr) {
		// check root allow list
		if isRootAllowed(addr) {
			var root model.User
			if err := model.DB.Select("id").Where("role = ?", model.RoleRootUser).First(&root).Error; err == nil {
				_ = root.FillUserById()
				root.WalletAddress = addr
				_ = root.Update(false)
				user = root
			}
		} else if config.WalletAutoRegisterEnabled {
			// auto create user
			username := "wallet_" + random.GetRandomString(6)
			for model.IsUsernameAlreadyTaken(username) {
				username = "wallet_" + random.GetRandomString(6)
			}
			user = model.User{
				Username:      username,
				Password:      random.GetRandomString(16),
				DisplayName:   username,
				Role:          model.RoleCommonUser,
				Status:        model.UserStatusEnabled,
				WalletAddress: addr,
			}
			_ = user.Insert(c.Request.Context(), 0)
		} else {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "未找到钱包绑定的账户，请先绑定或由管理员开启自动注册",
			})
			return
		}
	} else {
		if err := user.FillUserByWalletAddress(); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
		// 回收已删除账户占用的钱包地址
		if user.Status == model.UserStatusDeleted {
			_ = model.DB.Model(&user).Update("wallet_address", "")
			// 再次走注册/绑定逻辑
			if isRootAllowed(addr) {
				var root model.User
				if err := model.DB.Select("id").Where("role = ?", model.RoleRootUser).First(&root).Error; err == nil {
					_ = root.FillUserById()
					root.WalletAddress = addr
					_ = root.Update(false)
					user = root
				}
			} else if config.WalletAutoRegisterEnabled {
				username := "wallet_" + random.GetRandomString(6)
				for model.IsUsernameAlreadyTaken(username) {
					username = "wallet_" + random.GetRandomString(6)
				}
				user = model.User{
					Username:      username,
					Password:      random.GetRandomString(16),
					DisplayName:   username,
					Role:          model.RoleCommonUser,
					Status:        model.UserStatusEnabled,
					WalletAddress: addr,
				}
				_ = user.Insert(c.Request.Context(), 0)
			}
		}
	}

	if user.Status != model.UserStatusEnabled {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "用户已被封禁",
		})
		return
	}
	controller.SetupLogin(&user, c)
	common.ConsumeWalletNonce(addr)
}

// WalletBind binds a wallet to logged-in user
func WalletBind(c *gin.Context) {
	if !config.WalletLoginEnabled {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "管理员未开启钱包登录",
		})
		return
	}
	var req walletLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "参数错误",
		})
		return
	}
	if err := verifyWalletRequest(req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	addr := strings.ToLower(req.Address)
	session := sessions.Default(c)
	id := session.Get("id")
	if id == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "未登录",
		})
		return
	}
	user := model.User{Id: id.(int)}
	if err := user.FillUserById(); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	if model.IsWalletAddressAlreadyTaken(addr) {
		exist := model.User{WalletAddress: addr}
		if err := exist.FillUserByWalletAddress(); err == nil {
			if exist.Status == model.UserStatusDeleted {
				_ = model.DB.Model(&exist).Update("wallet_address", "")
			} else if strings.ToLower(user.WalletAddress) != addr && exist.Id != user.Id {
				c.JSON(http.StatusOK, gin.H{
					"success": false,
					"message": "该钱包已绑定其他账户",
				})
				return
			}
		}
	}
	user.WalletAddress = addr
	if err := user.Update(false); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	common.ConsumeWalletNonce(addr)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "绑定成功",
	})
}

func verifyWalletRequest(req walletLoginRequest) error {
	if !common.IsValidEthAddress(req.Address) {
		return errors.New("无效的钱包地址")
	}
	if req.Signature == "" || req.Nonce == "" {
		return errors.New("缺少签名或 nonce")
	}
	// chainId check
	if len(config.WalletAllowedChains) > 0 && req.ChainId != "" {
		allowed := false
		for _, c := range config.WalletAllowedChains {
			if strings.TrimSpace(c) == req.ChainId {
				allowed = true
				break
			}
		}
		if !allowed {
			return errors.New("不允许的链 ID")
		}
	}
	entry, ok := common.GetWalletNonce(req.Address)
	if !ok || entry.Nonce != req.Nonce {
		return errors.New("nonce 无效或已过期")
	}
	// verify signature
	recovered, err := recoverAddress(entry.Message, req.Signature)
	if err != nil {
		logger.SysError("wallet login verify failed: " + err.Error())
		return errors.New("签名验证失败")
	}
	if strings.ToLower(recovered) != strings.ToLower(req.Address) {
		return errors.New("签名地址与请求地址不一致")
	}
	return nil
}

func recoverAddress(message, signature string) (string, error) {
	sig := strings.TrimPrefix(signature, "0x")
	raw, err := hex.DecodeString(sig)
	if err != nil {
		return "", err
	}
	if len(raw) != 65 {
		return "", errors.New("签名长度异常")
	}
	// fix v value
	if raw[64] >= 27 {
		raw[64] -= 27
	}
	hash := accounts.TextHash([]byte(message))
	pub, err := crypto.SigToPub(hash, raw)
	if err != nil {
		return "", err
	}
	addr := crypto.PubkeyToAddress(*pub)
	return strings.ToLower(addr.Hex()), nil
}

func isRootAllowed(addr string) bool {
	for _, a := range config.WalletRootAllowedAddresses {
		if strings.ToLower(a) == addr {
			return true
		}
	}
	return false
}
