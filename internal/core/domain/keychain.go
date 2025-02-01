package domain

type KeyChainID uint64

type KCData struct {
	ID      KeyChainID
	OwnerID UserID
}

type KCItemType byte

const (
	KCItemTypePassword   KCItemType = iota + 1
	KCItemTypeString     KCItemType = iota + 1
	KCItemTypeBinary     KCItemType = iota + 1
	KCItemTypeCardNumber KCItemType = iota + 1
)

type KeyChainItemID uint64

const (
	KCMetaKeyComment = "Comment"
	KCMetaKeyLogin   = "Login"
	KCMetaKeySite    = "Site"
	KCMetaKeyIssuer  = "Issuer"
	KCMetaKeyOwner   = "Owner"
	KCMetaKeyValidTo = "ValidTo"
)

type KeyChainItemMeta map[string]string

type KCItemData struct {
	Label      string
	MetaData   KeyChainItemMeta
	Value      []byte
	Key        []byte
	KeyChainID KeyChainID
	ID         KeyChainItemID
	ItemType   KCItemType
}
