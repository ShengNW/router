package group

import billingratio "github.com/yeying-community/router/internal/relay/billing/ratio"

func List() []string {
	groupNames := make([]string, 0, len(billingratio.GroupRatio))
	for groupName := range billingratio.GroupRatio {
		groupNames = append(groupNames, groupName)
	}
	return groupNames
}
