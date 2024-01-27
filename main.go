package main

import (
	"github.com/A7mad-2000as/GoFish/chessEngine"
)

func main() {
	InitEvaluationRelatedMasks()
	chessEngine.InitializeLateMoveReductions()
	engineInterface := chessEngine.NewCustomEngineInterface(&chessEngine.DefaultSearcher{}, &CustomEvaluator{})
	engineInterface.StartEngine()
}
