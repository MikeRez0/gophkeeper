package domain

type KeyChainID uint64

type KeyChain struct {
	ID      KeyChainID
	OwnerID UserID
}

type KCItemType byte

const (
	KCItemTypePassword = iota
	KCItemTypeString
	KCItemTypeBinary
	KCItemTypeCardNumber
)

type KeyChainRecordID uint64

type KeyChainRecord struct {
	Label      string
	Login      string
	Comment    string
	Value      []byte
	ID         KeyChainRecordID
	KeyChainID KeyChainID
	Type       KCItemType
}

type KeyChainRecordMeta struct {
	Key        string
	Value      string
	ID         KeyChainRecordID
	KeyChainID KeyChainID
}
