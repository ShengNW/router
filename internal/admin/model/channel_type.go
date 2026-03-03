package model

// ChannelTypeCatalog stores channel interface type options for admin UI.
type ChannelTypeCatalog struct {
	ID          int    `json:"id" gorm:"primaryKey;autoIncrement:false"`
	Name        string `json:"name" gorm:"type:varchar(64);default:''"`
	Label       string `json:"label" gorm:"type:varchar(128);default:''"`
	Color       string `json:"color" gorm:"type:varchar(32);default:''"`
	Description string `json:"description" gorm:"type:text"`
	Tip         string `json:"tip" gorm:"type:text"`
	Source      string `json:"source" gorm:"type:varchar(32);default:'default'"`
	Enabled     bool   `json:"enabled" gorm:"default:true"`
	SortOrder   int    `json:"sort_order" gorm:"column:sort_order;default:0"`
	UpdatedAt   int64  `json:"updated_at" gorm:"bigint"`
}

func (ChannelTypeCatalog) TableName() string {
	return "channel_types"
}
