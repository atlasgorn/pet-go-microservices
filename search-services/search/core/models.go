package core

type DBStats struct {
	WordsTotal    int
	WordsUnique   int
	ComicsFetched int
}

type ServiceStats struct {
	DBStats
	ComicsTotal int
}

type Comics struct {
	ID    int
	URL   string
	Words []string
}

type PbComic struct {
	ID  int
	URL string
}
