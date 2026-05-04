package storage

import (
	"sync"
)

var SubtitleTasks = sync.Map{} // task id -> SubtitleTask, used for API data queries
