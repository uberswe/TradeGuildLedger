package client

func addLog(s string) {
	logData = append(logData, s)
	list.Refresh()
	list.Select(len(logData) - 1)
}
