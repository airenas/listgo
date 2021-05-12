package mongo

const (
	store        = "store"
	statusTable  = "status"
	resultTable  = "result"
	requestTable = "request"
	workTable    = "work"
	emailTable   = "emailLock"
)

var indexData = []IndexData{
	newIndexData(statusTable, "ID", true),
	newIndexData(resultTable, "ID", true),
	newIndexData(requestTable, "ID", true),
	newIndexData(emailTable, "ID", false),
	newIndexData(workTable, "ID", true),
}
