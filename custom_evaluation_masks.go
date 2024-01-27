package main

import (
	"github.com/A7mad-2000as/GoFish/chessEngine"
)

type KingSafetyZone struct {
	OuterDefenseRing chessEngine.Bitboard
	InnerDefenseRing chessEngine.Bitboard
}

var CheckForIsolatedPawnOnFileMasks [8]chessEngine.Bitboard
var CheckDoublePawnOnSquareMask [2][64]chessEngine.Bitboard
var CheckPassedPawnOnSquareMask [2][64]chessEngine.Bitboard
var CheckOutpostOnSquareMask [2][64]chessEngine.Bitboard
var KingSafetyZonesOnSquareMask [64]KingSafetyZone

func InitEvaluationRelatedMasks() {
	for file := chessEngine.FileA; file <= chessEngine.FileH; file++ {
		computeCheckIsolatedPawnOnFileMask(uint8(file))
	}
	for square := 0; square < 64; square++ {
		computeCheckDoublePawnOnSquareMask(uint8(square))
		computeOutpostOnSquareMask(uint8(square))
		computeKingSafetyZonesOnSquareMask(uint8(square))
		computeCheckPassedPawnOnSquareMask(uint8(square))
	}
}

func computeCheckIsolatedPawnOnFileMask(file uint8) {
	var bitboardForFile chessEngine.Bitboard = chessEngine.SetFileMasks[file]
	var CheckIsolatedPawnOnFileMask chessEngine.Bitboard = ((bitboardForFile & chessEngine.ClearFileMasks[chessEngine.FileA]) << 1) | ((bitboardForFile & chessEngine.ClearFileMasks[chessEngine.FileH]) >> 1)
	CheckForIsolatedPawnOnFileMasks[file] = CheckIsolatedPawnOnFileMask
}
func computeCheckDoublePawnOnSquareMask(square uint8) {
	var bitboardForFile chessEngine.Bitboard = chessEngine.SetFileMasks[chessEngine.File(square)]
	squareRank := int(chessEngine.Rank(square))
	var whiteMask chessEngine.Bitboard = bitboardForFile
	for rank := 0; rank <= squareRank; rank++ {
		whiteMask &= chessEngine.ClearRankMasks[rank]
	}
	CheckDoublePawnOnSquareMask[chessEngine.White][square] = whiteMask
	var blackMask chessEngine.Bitboard = bitboardForFile
	for rank := 7; rank >= squareRank; rank-- {
		blackMask &= chessEngine.ClearRankMasks[rank]
	}
	CheckDoublePawnOnSquareMask[chessEngine.Black][square] = blackMask

}
func computeCheckPassedPawnOnSquareMask(square uint8) {
	var currentFileAndAdjacentFilesMask chessEngine.Bitboard = CheckForIsolatedPawnOnFileMasks[chessEngine.File(square)] | chessEngine.SetFileMasks[chessEngine.File(square)]
	squareRank := int(chessEngine.Rank(square))
	whiteFrontSpanMask := currentFileAndAdjacentFilesMask
	for rank := 0; rank <= squareRank; rank++ {
		whiteFrontSpanMask &= chessEngine.ClearRankMasks[rank]
	}
	CheckPassedPawnOnSquareMask[chessEngine.White][square] = whiteFrontSpanMask

	blackFrontSpanMask := currentFileAndAdjacentFilesMask
	for rank := 7; rank >= squareRank; rank-- {
		blackFrontSpanMask &= chessEngine.ClearRankMasks[rank]
	}
	CheckPassedPawnOnSquareMask[chessEngine.Black][square] = blackFrontSpanMask

}
func computeOutpostOnSquareMask(square uint8) {
	var bitboardForFile chessEngine.Bitboard = chessEngine.SetFileMasks[chessEngine.File(square)]
	var currentAdjacentFilesMask chessEngine.Bitboard = CheckForIsolatedPawnOnFileMasks[chessEngine.File(square)]
	squareRank := int(chessEngine.Rank(square))
	whiteMask := currentAdjacentFilesMask
	for rank := 0; rank <= squareRank; rank++ {
		whiteMask &= chessEngine.ClearRankMasks[rank]
	}

	CheckOutpostOnSquareMask[chessEngine.White][square] = whiteMask & ^bitboardForFile

	blackMask := currentAdjacentFilesMask
	for rank := 7; rank >= squareRank; rank-- {
		blackMask &= chessEngine.ClearRankMasks[rank]
	}
	CheckOutpostOnSquareMask[chessEngine.Black][square] = blackMask & ^bitboardForFile
}
func computeKingSafetyZonesOnSquareMask(square uint8) {
	squareBitboard := chessEngine.BitboardForSquare[square]
	var aroundKingZone chessEngine.Bitboard = ((squareBitboard & chessEngine.ClearFileMasks[chessEngine.FileH]) >> 1) | ((squareBitboard & (chessEngine.ClearFileMasks[chessEngine.FileG] & chessEngine.ClearFileMasks[chessEngine.FileH])) >> 2)
	aroundKingZone |= ((squareBitboard & chessEngine.ClearFileMasks[chessEngine.FileA]) << 1) | ((squareBitboard & (chessEngine.ClearFileMasks[chessEngine.FileB] & chessEngine.ClearFileMasks[chessEngine.FileA])) << 2)
	aroundKingZone |= squareBitboard
	aroundKingZone |= (aroundKingZone >> 8) | (aroundKingZone >> 16)
	aroundKingZone |= (aroundKingZone << 8) | (aroundKingZone << 16)

	outerRing := aroundKingZone & ^(chessEngine.ComputedKingMoves[square] | squareBitboard)
	innerRing := chessEngine.ComputedKingMoves[square] | squareBitboard
	KingSafetyZonesOnSquareMask[square] = KingSafetyZone{
		OuterDefenseRing: outerRing,
		InnerDefenseRing: innerRing,
	}
}
