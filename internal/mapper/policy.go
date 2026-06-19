package mapper

import (
	p "reverseproxy/internal/domain/policy"
	dto "reverseproxy/internal/dto/policy"
)

func ToPolicyListItems(items []p.PolicyListItem) []dto.ListItem {
	listItems := make([]dto.ListItem, 0, len(items))
	for _, item := range items {
		listItems = append(listItems, dto.ListItem{
			ID:      item.ID.Hex(),
			Name:    item.Name,
			WL:      item.WL,
			Webapps: item.Webapps,
		})
	}
	return listItems
}

func ToPolicyDetail(d *p.Detail) dto.Detail {
	rules := make([]dto.RuleRefView, 0, len(d.Rules))

	for _, r := range d.Rules {
		rules = append(rules, dto.RuleRefView{
			ID:      r.ID.Hex(),
			Name:    r.Name,
			Enabled: r.Enabled,
		})
	}

	return dto.Detail{
		ID:    d.ID.Hex(),
		Name:  d.Name,
		WL:    d.WL,
		Rules: rules,
	}
}
