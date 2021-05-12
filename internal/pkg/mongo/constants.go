package mongo

const (
	store        = "store"
	statusTable  = "status"
	resultTable  = "result"
	requestTable = "request"
	relatedTable = "related"
	emailTable   = "emailLock"
)

var indexData = []IndexData{
	newIndexData(statusTable, "ID", true),
	newIndexData(resultTable, "ID", true),
	newIndexData(requestTable, "ID", true),
	newIndexData(emailTable, "ID", false),
	newIndexData(relatedTable, "ID", true),
}
