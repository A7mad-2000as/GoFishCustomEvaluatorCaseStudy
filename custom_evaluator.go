package main

import (
	"github.com/A7mad-2000as/GoFish/chessEngine"
)

const (
	CheckmateScore             int16 = 10000
	drawScore                  int16 = 0
	DrawishPositionScaleFactor int16 = 16
)

type CustomEvaluator struct {
	evaluationData EvaluationData
}

type EvaluationData struct {
	MidgameScores           [2]int16
	EndgameScores           [2]int16
	ThreatToEnemyKingPoints [2]uint16
	EnemyKingAttackerCount  [2]uint8
}

func (evaluator *CustomEvaluator) GetMiddleGamePieceSquareTable() *[6][64]int16 {
	return &MidGamePieceSquareTables
}

func (evaluator *CustomEvaluator) GetEndGamePieceSquareTable() *[6][64]int16 {
	return &EndGamePieceSquareTables
}

func (evaluator *CustomEvaluator) GetMiddleGamePieceValues() *[6]int16 {
	return &MidGamePieceValues
}

func (evaluator *CustomEvaluator) GetEndGamePieceValues() *[6]int16 {
	return &EndGamePieceValues
}

func (evaluator *CustomEvaluator) GetPhaseValues() *[6]int16 {
	return &PiecePhaseIncrements
}

func (evaluator *CustomEvaluator) GetTotalPhaseWeight() int16 {
	return TotalPhaseIncrement
}

func (customClassicEvaluator *CustomEvaluator) EvaluatePosition(position *chessEngine.Position) int16 {
	if isDrawnState(position) {
		return drawScore
	}
	allBitBoard := position.ColorsBitBoard[position.SideToMove] | position.ColorsBitBoard[position.SideToMove^1]
	var phaseValue = position.Phase
	customClassicEvaluator.evaluationData = EvaluationData{
		MidgameScores: position.MidGameScores,
		EndgameScores: position.EndGameScores,
	}
	for allBitBoard != 0 {
		pieceSquare := allBitBoard.PopMostSignificantBit()
		pieceType := position.SquareContent[pieceSquare].PieceType
		pieceColor := position.SquareContent[pieceSquare].Color
		switch pieceType {
		case chessEngine.Pawn:
			customClassicEvaluator.evaluatePawnAtSquare(position, pieceColor, pieceSquare)
		case chessEngine.Knight:
			customClassicEvaluator.evaluateKnightAtSquare(position, pieceColor, pieceSquare)
		case chessEngine.Bishop:
			customClassicEvaluator.evaluateBishopAtSquare(position, pieceColor, pieceSquare)
		case chessEngine.Rook:
			customClassicEvaluator.evaluateRookAtSquare(position, pieceColor, pieceSquare)
		case chessEngine.Queen:
			customClassicEvaluator.evaluateQueenAtSquare(position, pieceColor, pieceSquare)
		}
	}
	for color := chessEngine.Black; color <= chessEngine.White; color++ {
		if position.PiecesBitBoard[color][chessEngine.Bishop].CountSetBits() >= 2 {
			customClassicEvaluator.evaluationData.MidgameScores[color] += MidGameBishopPairBonus
			customClassicEvaluator.evaluationData.EndgameScores[color] += EndgameBishopPairBonus
		}
		customClassicEvaluator.evaluateKingAtSquare(position, color, position.PiecesBitBoard[color][chessEngine.King].MostSignificantBit())
	}
	customClassicEvaluator.evaluationData.MidgameScores[position.SideToMove] += MidGameTempoBonus

	currentMidGameScore := customClassicEvaluator.evaluationData.MidgameScores[position.SideToMove] - customClassicEvaluator.evaluationData.MidgameScores[position.SideToMove^1]
	currentEndGameScore := customClassicEvaluator.evaluationData.EndgameScores[position.SideToMove] - customClassicEvaluator.evaluationData.EndgameScores[position.SideToMove^1]

	scaledPhaseValue := (phaseValue*256 + (TotalPhaseIncrement / 2)) / TotalPhaseIncrement
	currentScore := int16(((int32(currentMidGameScore) * (int32(256) - int32(scaledPhaseValue))) + (int32(currentEndGameScore) * int32(scaledPhaseValue))) / int32(256))

	if isDrawishState(position) {
		return currentScore / DrawishPositionScaleFactor
	}

	return currentScore
}
func (customClassicEvaluator *CustomEvaluator) evaluatePawnAtSquare(position *chessEngine.Position, color uint8, square uint8) {
	enemyPawns := position.PiecesBitBoard[color^1][chessEngine.Pawn]
	sideToMovePawn := position.PiecesBitBoard[color][chessEngine.Pawn]
	fileOfSq := chessEngine.File(square)
	isIsolated := CheckForIsolatedPawnOnFileMasks[fileOfSq]&sideToMovePawn == 0
	isDoubled := CheckDoublePawnOnSquareMask[color][square]&sideToMovePawn != 0
	isPassedAndNotBlockedByFriendlyPawn := CheckPassedPawnOnSquareMask[color][square]&enemyPawns == 0 && sideToMovePawn&CheckDoublePawnOnSquareMask[color][square] == 0

	if isIsolated {
		customClassicEvaluator.evaluationData.MidgameScores[color] -= MidGameIsolatedPawnPenalty
		customClassicEvaluator.evaluationData.EndgameScores[color] -= EndGameIsolatedPawnPenalty
	}
	if isDoubled {
		customClassicEvaluator.evaluationData.MidgameScores[color] -= MidGameDoubledPawnPenalty
		customClassicEvaluator.evaluationData.EndgameScores[color] -= EndGameDoubledPawnPenalty
	}
	if isPassedAndNotBlockedByFriendlyPawn {
		customClassicEvaluator.evaluationData.MidgameScores[color] += MidGamePassedPawnSquareTables[chessEngine.BoardSquaresNormalAndFlipped[color][square]]
		customClassicEvaluator.evaluationData.EndgameScores[color] += EndGamePassedPawnSquareTables[chessEngine.BoardSquaresNormalAndFlipped[color][square]]
	}
}
func (customClassicEvaluator *CustomEvaluator) evaluateKnightAtSquare(position *chessEngine.Position, color uint8, square uint8) {
	var enemyPawns chessEngine.Bitboard = position.PiecesBitBoard[color^1][chessEngine.Pawn]
	var sideToMovePawns chessEngine.Bitboard = position.PiecesBitBoard[color][chessEngine.Pawn]

	// outPost Evaluation
	noEnemyCanAttackKnight := CheckOutpostOnSquareMask[color][square]&enemyPawns == 0
	isTheKnightProtectedByFriendlyPawn := chessEngine.ComputedPawnCaptures[color^1][square]&sideToMovePawns != 0
	if noEnemyCanAttackKnight && isTheKnightProtectedByFriendlyPawn &&
		chessEngine.BoardRanksNormalAndFlipped[color][chessEngine.Rank(square)] >= chessEngine.Rank5 {
		customClassicEvaluator.evaluationData.MidgameScores[color] += MidGameKnightOnOutpostBonus
		customClassicEvaluator.evaluationData.EndgameScores[color] += EndGameKnightOnOutpostBonus
	}
	// mobility evaluation
	var sideToMoveBitBoard chessEngine.Bitboard = position.ColorsBitBoard[color]
	var knightMoves chessEngine.Bitboard = chessEngine.ComputedKnightMoves[square] & ^sideToMoveBitBoard
	var knightSafeMoves chessEngine.Bitboard = filterMoveAndKeepTheSafeMoves(knightMoves, color, enemyPawns)
	mobility := int16(knightSafeMoves.CountSetBits())
	customClassicEvaluator.evaluationData.MidgameScores[color] += (mobility - 4) * MidGameMobilityScoresPerPiece[chessEngine.Knight]
	customClassicEvaluator.evaluationData.EndgameScores[color] += (mobility - 4) * EndGameMobilityScoresPerPiece[chessEngine.Knight]

	// attacks on enemy king evaluation
	customClassicEvaluator.evaluateAttacksOnEnemyKing(position, knightSafeMoves, color, chessEngine.Knight)
}
func (customClassicEvaluator *CustomEvaluator) evaluateBishopAtSquare(position *chessEngine.Position, color uint8, square uint8) {
	enemyPawns := position.PiecesBitBoard[color^1][chessEngine.Pawn]
	sideToMovePawns := position.PiecesBitBoard[color][chessEngine.Pawn]
	sideToMoveBitBoard := position.ColorsBitBoard[color]
	allBitBoard := position.ColorsBitBoard[position.SideToMove] | position.ColorsBitBoard[position.SideToMove^1]
	// outPost Evaluation
	noEnemyCanAttackBishop := CheckOutpostOnSquareMask[color][square]&enemyPawns == 0
	isTheBishopProtectedByFriendlyPawn := chessEngine.ComputedPawnCaptures[color^1][square]&sideToMovePawns != 0
	if noEnemyCanAttackBishop && isTheBishopProtectedByFriendlyPawn &&
		chessEngine.BoardRanksNormalAndFlipped[color][chessEngine.Rank(square)] >= chessEngine.Rank5 {
		customClassicEvaluator.evaluationData.MidgameScores[color] += MidGameBishopOnOutpostBonus
		customClassicEvaluator.evaluationData.EndgameScores[color] += EndGameBishopOnOutpostBonus
	}

	//mobility evaluation
	var bishopMoves chessEngine.Bitboard = chessEngine.GetBishopPseudoLegalMoves(square, allBitBoard) & ^sideToMoveBitBoard
	mobility := int16(bishopMoves.CountSetBits())
	customClassicEvaluator.evaluationData.MidgameScores[color] += (mobility - 7) * MidGameMobilityScoresPerPiece[chessEngine.Bishop]
	customClassicEvaluator.evaluationData.EndgameScores[color] += (mobility - 7) * EndGameMobilityScoresPerPiece[chessEngine.Bishop]

	// attacks on enemy king evaluation
	customClassicEvaluator.evaluateAttacksOnEnemyKing(position, bishopMoves, color, chessEngine.Bishop)

}
func (customClassicEvaluator *CustomEvaluator) evaluateRookAtSquare(position *chessEngine.Position, color uint8, square uint8) {
	enemyKingSquare := position.PiecesBitBoard[color^1][chessEngine.King].MostSignificantBit()
	allPawns := position.PiecesBitBoard[color][chessEngine.Pawn] | position.PiecesBitBoard[color^1][chessEngine.Pawn]
	sideToMoveBitBoard := position.ColorsBitBoard[color]
	allBitBoard := position.ColorsBitBoard[color] | position.ColorsBitBoard[color^1]

	if chessEngine.BoardRanksNormalAndFlipped[color][chessEngine.Rank(square)] == chessEngine.Rank7 && chessEngine.BoardRanksNormalAndFlipped[color][chessEngine.Rank(enemyKingSquare)] >= chessEngine.Rank7 {
		customClassicEvaluator.evaluationData.EndgameScores[color] += EndGameBonusForRookOrQueenOnSeventhRank
	}

	if chessEngine.SetFileMasks[chessEngine.File(square)]&allPawns == 0 {
		customClassicEvaluator.evaluationData.MidgameScores[color] += MidGameRookOnOpenFileBonus
	}

	rookMoves := chessEngine.GetRookPseudoLegalMoves(square, allBitBoard) & ^sideToMoveBitBoard
	mobility := int16(rookMoves.CountSetBits())
	customClassicEvaluator.evaluationData.MidgameScores[color] += (mobility - 7) * MidGameMobilityScoresPerPiece[chessEngine.Rook]
	customClassicEvaluator.evaluationData.EndgameScores[color] += (mobility - 7) * EndGameMobilityScoresPerPiece[chessEngine.Rook]
	customClassicEvaluator.evaluateAttacksOnEnemyKing(position, rookMoves, color, chessEngine.Rook)

}
func (customClassicEvaluator *CustomEvaluator) evaluateQueenAtSquare(position *chessEngine.Position, color uint8, square uint8) {
	enemyKingSquare := position.PiecesBitBoard[color^1][chessEngine.King].MostSignificantBit()
	sideToMoveBitBoard := position.ColorsBitBoard[color]
	allBitBoard := position.ColorsBitBoard[color] | position.ColorsBitBoard[color^1]

	if chessEngine.BoardRanksNormalAndFlipped[color][chessEngine.Rank(square)] == chessEngine.Rank7 && chessEngine.BoardRanksNormalAndFlipped[color][chessEngine.Rank(enemyKingSquare)] >= chessEngine.Rank7 {
		customClassicEvaluator.evaluationData.EndgameScores[color] += EndGameBonusForRookOrQueenOnSeventhRank
	}
	queenMoves := (chessEngine.GetBishopPseudoLegalMoves(square, allBitBoard) | chessEngine.GetRookPseudoLegalMoves(square, allBitBoard)) & ^sideToMoveBitBoard
	mobility := int16(queenMoves.CountSetBits())

	customClassicEvaluator.evaluationData.MidgameScores[color] += (mobility - 14) * MidGameMobilityScoresPerPiece[chessEngine.Queen]
	customClassicEvaluator.evaluationData.EndgameScores[color] += (mobility - 14) * EndGameMobilityScoresPerPiece[chessEngine.Queen]

	customClassicEvaluator.evaluateAttacksOnEnemyKing(position, queenMoves, color, chessEngine.Queen)
}

func (customClassicEvaluator *CustomEvaluator) evaluateKingAtSquare(position *chessEngine.Position, color uint8, square uint8) {
	threatPointOnSideToMoveKing := customClassicEvaluator.evaluationData.ThreatToEnemyKingPoints[color^1]
	kingFile := chessEngine.SetFileMasks[chessEngine.File(square)]
	kingLeftFile, kingRightFile := ((kingFile & chessEngine.ClearFileMasks[chessEngine.FileA]) << 1), ((kingFile & chessEngine.ClearFileMasks[chessEngine.FileH]) >> 1)
	sideToMovePawns := position.PiecesBitBoard[color][chessEngine.Pawn]

	var semipOpenFilePenality uint16 = 0
	if kingFile&sideToMovePawns == 0 {
		semipOpenFilePenality += uint16(SemiOpenFileBesideKingPenalty)
	}
	if kingLeftFile != 0 && kingLeftFile&sideToMovePawns == 0 {
		semipOpenFilePenality += uint16(SemiOpenFileBesideKingPenalty)
	}
	if kingRightFile != 0 && kingRightFile&sideToMovePawns == 0 {
		semipOpenFilePenality += uint16(SemiOpenFileBesideKingPenalty)
	}

	finalPenalty := int16(((threatPointOnSideToMoveKing + semipOpenFilePenality) * (threatPointOnSideToMoveKing + semipOpenFilePenality)) / 4)
	if customClassicEvaluator.evaluationData.EnemyKingAttackerCount[color^1] >= 2 && position.PiecesBitBoard[color^1][chessEngine.Queen] != 0 {
		customClassicEvaluator.evaluationData.MidgameScores[color] -= finalPenalty
	}
}

func (customClassicEvaluator *CustomEvaluator) evaluateAttacksOnEnemyKing(position *chessEngine.Position, moves chessEngine.Bitboard, color uint8, piece uint8) {
	var attacksOnEnemyKingOuterRing chessEngine.Bitboard = moves & KingSafetyZonesOnSquareMask[position.PiecesBitBoard[color^1][chessEngine.King].MostSignificantBit()].OuterDefenseRing
	var attacksOnEnemyKingInnerRing chessEngine.Bitboard = moves & KingSafetyZonesOnSquareMask[position.PiecesBitBoard[color^1][chessEngine.King].MostSignificantBit()].InnerDefenseRing
	if attacksOnEnemyKingOuterRing != 0 || attacksOnEnemyKingInnerRing != 0 {
		customClassicEvaluator.evaluationData.EnemyKingAttackerCount[color]++
		customClassicEvaluator.evaluationData.ThreatToEnemyKingPoints[color] += uint16(attacksOnEnemyKingOuterRing.CountSetBits()) * uint16(OuterRingAttackScorePerPiece[piece])
		customClassicEvaluator.evaluationData.ThreatToEnemyKingPoints[color] += uint16(attacksOnEnemyKingInnerRing.CountSetBits()) * uint16(InnerRingAttackScorePerPiece[piece])
	}
}

func filterMoveAndKeepTheSafeMoves(moves chessEngine.Bitboard, color uint8, enemyPawns chessEngine.Bitboard) chessEngine.Bitboard {
	safeMoves := moves
	for enemyPawns != 0 {
		square := enemyPawns.PopMostSignificantBit()
		safeMoves &= ^chessEngine.ComputedPawnCaptures[color^1][square]
	}
	return safeMoves
}

func isDrawnState(position *chessEngine.Position) bool {
	whiteKnightCount := position.PiecesBitBoard[chessEngine.White][chessEngine.Knight].CountSetBits()
	whiteBishopCount := position.PiecesBitBoard[chessEngine.White][chessEngine.Bishop].CountSetBits()

	blackKnightCount := position.PiecesBitBoard[chessEngine.Black][chessEngine.Knight].CountSetBits()
	blackBishopCount := position.PiecesBitBoard[chessEngine.Black][chessEngine.Bishop].CountSetBits()

	totalPawnsCount := position.PiecesBitBoard[chessEngine.White][chessEngine.Pawn].CountSetBits() + position.PiecesBitBoard[chessEngine.Black][chessEngine.Pawn].CountSetBits()
	totalKnightsCount := whiteKnightCount + blackKnightCount
	totalBishopsCount := whiteBishopCount + blackBishopCount
	totalRooksCount := position.PiecesBitBoard[chessEngine.White][chessEngine.Rook].CountSetBits() + position.PiecesBitBoard[chessEngine.Black][chessEngine.Rook].CountSetBits()
	totalQueensCount := position.PiecesBitBoard[chessEngine.White][chessEngine.Queen].CountSetBits() + position.PiecesBitBoard[chessEngine.Black][chessEngine.Queen].CountSetBits()

	majorPiecesCount := totalRooksCount + totalQueensCount
	minorPiecesCount := totalKnightsCount + totalBishopsCount

	if totalPawnsCount+majorPiecesCount+minorPiecesCount == 0 {
		// King vs King
		return true
	} else if majorPiecesCount+totalPawnsCount == 0 {
		if minorPiecesCount == 1 {
			// King & minorPiece vs King
			return true
		} else if minorPiecesCount == 2 && whiteKnightCount == 1 && blackKnightCount == 1 {
			// King & Knight vs King & Knight
			return true
		} else if minorPiecesCount == 2 && whiteBishopCount == 1 && blackBishopCount == 1 {
			// King & Bishop vs King & Bishop and bishops are on the same square color
			return isSquareLight(position.PiecesBitBoard[chessEngine.White][chessEngine.Bishop].MostSignificantBit()) == isSquareLight(position.PiecesBitBoard[chessEngine.Black][chessEngine.Bishop].MostSignificantBit())
		}
	}

	return false
}

func isDrawishState(pos *chessEngine.Position) bool {
	whiteKnightCount := pos.PiecesBitBoard[chessEngine.White][chessEngine.Knight].CountSetBits()
	whiteBishopCount := pos.PiecesBitBoard[chessEngine.White][chessEngine.Bishop].CountSetBits()
	whiteRookCount := pos.PiecesBitBoard[chessEngine.White][chessEngine.Rook].CountSetBits()
	whiteQueenCount := pos.PiecesBitBoard[chessEngine.White][chessEngine.Queen].CountSetBits()

	blackKnightCount := pos.PiecesBitBoard[chessEngine.Black][chessEngine.Knight].CountSetBits()
	blackBishopCount := pos.PiecesBitBoard[chessEngine.Black][chessEngine.Bishop].CountSetBits()
	blackRookCount := pos.PiecesBitBoard[chessEngine.Black][chessEngine.Rook].CountSetBits()
	blackQueenCount := pos.PiecesBitBoard[chessEngine.Black][chessEngine.Queen].CountSetBits()

	totalPawnsCount := pos.PiecesBitBoard[chessEngine.White][chessEngine.Pawn].CountSetBits() + pos.PiecesBitBoard[chessEngine.Black][chessEngine.Pawn].CountSetBits()
	totalKnightsCount := whiteKnightCount + blackKnightCount
	totalBishopsCount := whiteBishopCount + blackKnightCount
	totalRooksCount := whiteRookCount + blackRookCount
	totalQueensCount := whiteQueenCount + blackQueenCount

	whiteMinorPiecesCount := whiteBishopCount + whiteKnightCount
	blackMinorPiecesCount := blackBishopCount + blackKnightCount

	totalMajorPiecesCount := totalRooksCount + totalQueensCount
	totalMinorPiecesCount := totalKnightsCount + totalBishopsCount
	totalPiecesCount := totalMajorPiecesCount + totalMinorPiecesCount

	if totalPawnsCount == 0 {
		if totalPiecesCount == 2 && blackQueenCount == 1 && whiteQueenCount == 1 {
			// KQ v KQ
			return true
		} else if totalPiecesCount == 2 && blackRookCount == 1 && whiteRookCount == 1 {
			// KR v KR
			return true
		} else if totalPiecesCount == 2 && whiteMinorPiecesCount == 1 && blackMinorPiecesCount == 1 {
			// KN v KB
			// KB v KB
			return true
		} else if totalPiecesCount == 3 && ((whiteQueenCount == 1 && blackRookCount == 2) || (blackQueenCount == 1 && whiteRookCount == 2)) {
			// KQ v KRR
			return true
		} else if totalPiecesCount == 3 && ((whiteQueenCount == 1 && blackBishopCount == 2) || (blackQueenCount == 1 && whiteBishopCount == 2)) {
			// KQ vs KBB
			return true
		} else if totalPiecesCount == 3 && ((whiteQueenCount == 1 && blackKnightCount == 2) || (blackQueenCount == 1 && whiteKnightCount == 2)) {
			// KQ vs KNN
			return true
		} else if totalPiecesCount <= 3 && ((whiteKnightCount == 2 && blackMinorPiecesCount <= 1) || (blackKnightCount == 2 && whiteMinorPiecesCount <= 1)) {
			// KNN v KN, KNN v KB, KNN v K
			return true
		} else if totalPiecesCount == 3 &&
			((whiteQueenCount == 1 && blackRookCount == 1 && blackMinorPiecesCount == 1) ||
				(blackQueenCount == 1 && whiteRookCount == 1 && whiteMinorPiecesCount == 1)) {
			// KQ vs KRN, KQ vs KRB
			return true
		} else if totalPiecesCount == 3 &&
			((whiteRookCount == 1 && blackRookCount == 1 && blackMinorPiecesCount == 1) ||
				(blackRookCount == 1 && whiteRookCount == 1 && whiteMinorPiecesCount == 1)) {
			// KR vs KRB, KR vs KRN
		} else if totalPiecesCount == 4 &&
			((whiteRookCount == 2 && blackRookCount == 1 && blackMinorPiecesCount == 1) ||
				(blackRookCount == 2 && whiteRookCount == 1 && whiteMinorPiecesCount == 1)) {
			// KRR v KRB, KRR v KRN
			return true
		}
	}

	return false
}

func isSquareLight(square uint8) bool {
	fileNumber := chessEngine.File(square)
	rankNumber := chessEngine.File(square)
	return (fileNumber+rankNumber)%2 != 0
}
