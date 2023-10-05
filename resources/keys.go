package resources

func NewStringKey(id string, resourceType ResourceType) Key {
	return Key{
		ID:   id,
		Type: resourceType,
	}
}
