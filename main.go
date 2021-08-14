package main

func main() {
}

func NewDefaultFileHandlers(
	reqCreator *CollectingRequestCreator,
	repo *Repository,
) FileHandlers {
	return FileHandlers{
		repo.Save,
		reqCreator.CreateCollectingRequestForFolder,
	}
}
