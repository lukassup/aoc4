package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func timeit(start time.Time, name string) {
	elapsed := time.Since(start)
	fmt.Printf("# %s duration: %+v\n", name, elapsed)
}

const boardSize = 5

type board [boardSize][boardSize]int

func printBoard(board board) {
	for _, row := range board {
		var str string
		for pos, val := range row {
			if pos > 0 {
				str += fmt.Sprintf(",%3d", val)
			} else {
				str += fmt.Sprintf("%3d", val)
			}
		}
		fmt.Println(str)
	}
}

func parseNumberDraws(scanner *bufio.Scanner) (numbers []int) {
	defer timeit(time.Now(), "parseNumberDraws")
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) > 0 && strings.Contains(line, ",") {
			for _, numstring := range strings.Split(line, ",") {
				number, err := strconv.Atoi(numstring)
				check(err)
				numbers = append(numbers, number)
			}
			break
		}
	}
	err := scanner.Err()
	check(err)
	return
}

func parseNumberBoards(scanner *bufio.Scanner) (boards []board) {
	defer timeit(time.Now(), "parseNumberBoards")
	boards = []board{}
	var currentBoard board
	var currentRow int = 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) > 0 {
			// skip number draws line
			if strings.Contains(line, ",") {
				continue
			}
			if currentRow == 0 {
				currentBoard = board{}
			}
			for pos, numstring := range strings.Fields(line) {
				num, err := strconv.Atoi(numstring)
				check(err)
				currentBoard[currentRow][pos] = num
			}
			if currentRow < boardSize-1 {
				currentRow++
			} else {
				boards = append(boards, currentBoard)
				currentRow = 0
			}
		}
	}
	err := scanner.Err()
	check(err)
	return
}

func markDrawnNumber(boards []board, number int) []board {
	// set guessed numbers to -1 so we could filter only positive values when
	// calculating board score (sum)
	for b := range boards {
		for y := range boards[b] {
			for x := range boards[b][y] {
				if boards[b][y][x] == number {
					boards[b][y][x] = -1
				}
			}
		}
	}
	return boards
}

func findWinningBoards(boards []board) (winningBoards []board) {
	// since we set guessed numbers to -1:
	// - find any boards with row sum == -(boardSize)
	// - find are boards with column sum == -(boardSize)
	for _, board := range boards {
		winning := false
		// sum rows
		for _, row := range board {
			sum := 0
			for _, val := range row {
				sum += val
			}
			if sum == -boardSize {
				winningBoards = append(winningBoards, board)
				winning = true
				break
			}
		}
		if winning {
			continue
		}
		// sum columns
		for x := 0; x < boardSize; x++ {
			sum := 0
			for y := 0; y < boardSize; y++ {
				sum += board[y][x]
			}
			if sum == -boardSize {
				winningBoards = append(winningBoards, board)
				break
			}
		}
	}
	return
}

func calcBoardScore(board board) (score int) {
	// - sum all numbers on the board
	// - skip guessed (negative) numbers
	for _, row := range board {
		for _, val := range row {
			if val > 0 {
				score += val
			}
		}
	}
	return
}

func findHighestScoringBoard(boards []board) (bestBoard board) {
	// in case there is more than one board, pick the better one
	bestScore := 0
	for _, board := range boards {
		score := calcBoardScore(board)
		if score > bestScore {
			bestScore = score
			bestBoard = board
		}
	}
	return
}

func playBingoBestChoice(boards []board, numbers []int) (score int) {
	defer timeit(time.Now(), "playBingoBestChoice")
	for draw, currentNumber := range numbers {
		boards = markDrawnNumber(boards, currentNumber)
		winningBoards := findWinningBoards(boards)
		if len(winningBoards) > 0 {
			fmt.Printf(
				"draw #%02d, number: %d - found %d winning board(s)\n",
				draw+1, currentNumber, len(winningBoards))
			bestBoard := findHighestScoringBoard(winningBoards)
			printBoard(bestBoard)
			score = calcBoardScore(bestBoard) * currentNumber
			break
		}
	}
	return
}

func findNonWinningBoards(boards []board) (nonWinningBoards []board) {
	// since we set guessed numbers to -1:
	// - find all boards with no rows sum == -(boardSize)
	// - find all boards with no columns sum == -(boardSize)
	for _, board := range boards {
		winning := false
		// sum rows
		for _, row := range board {
			sum := 0
			for _, val := range row {
				sum += val
			}
			if sum == -boardSize {
				winning = true
				break
			}
		}
		if winning {
			continue
		}
		// sum columns
		for x := 0; x < boardSize; x++ {
			sum := 0
			for y := 0; y < boardSize; y++ {
				sum += board[y][x]
			}
			if sum == -boardSize {
				winning = true
				break
			}
		}
		if !winning {
			nonWinningBoards = append(nonWinningBoards, board)
		}
	}
	return
}

func playBingoWorstChoice(boards []board, numbers []int) (score int) {
	defer timeit(time.Now(), "playBingoWorstChoice")
	// select the board to win LAST
	// filter all winning boards until there's only one board left
	for draw, currentNumber := range numbers {
		boards = markDrawnNumber(boards, currentNumber)
		if len(boards) > 2 {
			// no longer need to iterate over boards that have already won
			boards = findNonWinningBoards(boards)
		} else {
			boards = findWinningBoards(boards)
			fmt.Printf(
				"draw #%02d, number: %2d - found %d last winning board(s)\n",
				draw+1, currentNumber, len(boards))
			printBoard(boards[0])
			score = calcBoardScore(boards[0]) * currentNumber
			break
		}
	}
	return
}

func main() {
	defer timeit(time.Now(), "main")
	filename := "input"
	if len(os.Args) > 1 {
		filename = os.Args[1]
	}

	fd, err := os.Open(filename)
	defer fd.Close()
	check(err)

	scanner := bufio.NewScanner(fd)
	numbers := parseNumberDraws(scanner)
	boards := parseNumberBoards(scanner)

	result1 := playBingoBestChoice(boards, numbers)
	fmt.Printf("part1 result: %+v\n", result1)

	fd.Seek(0, io.SeekStart)

	result2 := playBingoWorstChoice(boards, numbers)
	fmt.Printf("part2 result: %+v\n", result2)
}
