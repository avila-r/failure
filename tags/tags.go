package tags

type Tags map[string]string

func Merge(tags Tags, on *Tags) {
	if on == nil {
		return
	}

	if *on == nil {
		*on = make(Tags)
	}

	for k, v := range tags {
		if _, exists := (*on)[k]; !exists {
			(*on)[k] = v
		}
	}
}
