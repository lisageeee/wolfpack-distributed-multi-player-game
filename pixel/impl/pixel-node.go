package impl

import (
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"../../geometry"
	"../../shared"
	_ "image/png"
	_ "image/jpeg"
	"net"
	"fmt"
	"encoding/json"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/font/basicfont"
	"sort"
	"os"
	"image/color"
)

var NodeAddr string // must store as global to get it into run function
var MyAddr string

// Sprite size
const spriteStep = 30

type PixelNode struct {
	// The TCP connection with the associated logic node
	Sender            *net.TCPConn

	// This player's current position
	playerPosition    shared.Coord

	// The PixelManager used to convert incoming coordinates to Pixel-understandable vectors
	Geom              geometry.PixelManager

	// Incoming game states that have not yet been rendered
	NewGameStates     chan shared.GameRenderState

	// The sprite (graphic) that represents the player
	PlayerSprite      *pixel.Sprite

	// The sprite (graphic) that represents the walls
	WallSprite        *pixel.Sprite

	// The sprite (graphic) that represents the other players
	OtherPlayerSprite *pixel.Sprite

	// The sprite (graphic) that represents the prey
	PreySprite        *pixel.Sprite

	// The sprite (graphic) that represents scoreboard background
	ScoreboardBg 	  *imdraw.IMDraw

	// The text atlas which is required to draw text with pixel
	TextAtlas  		  *text.Atlas
}

// Creates a pixel node by setting up the TCP connection with the logic node, and getting the associated game settings.
// Returns the created pixel node.
func CreatePixelNode(nodeAddr string) (PixelNode) {

	// Setup connection
	remote := setupTCP(nodeAddr)

	// Get initial game state (first message after tcp setup)
	var buf = make([]byte, 2048)
	var settings shared.InitialGameSettings

	// Get initial game state
	rlen, err := remote.Read(buf)
	if err != nil {
		fmt.Println(err)
	} else {
		json.Unmarshal(buf[0:rlen], &settings)
	}

	// Init walls
	wallCoords := settings.WallCoordinates

	// Create geometry manager
	geom := geometry.CreatePixelManager(settings.WindowsX, settings.WindowsY, settings.ScoreboardWidth,
		spriteStep, wallCoords)

	// Create scoreboard
	scoreboardBg := createScoreboard(settings.WindowsX, settings.WindowsY, settings.ScoreboardWidth)

	// Allow text rendering
	basicAtlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)

	node := PixelNode{ Sender: remote, Geom: geom, NewGameStates: make(chan shared.GameRenderState, 5),
	ScoreboardBg: scoreboardBg, TextAtlas: basicAtlas}

	return node
}

//
func (pn * PixelNode) RenderNewState (win * pixelgl.Window, curState shared.GameRenderState) {


	// Clear current render
	win.Clear(color.RGBA{0x2d, 0x2d, 0x2d, 0xff})

	// Render walls, first
	pn.DrawWalls(win)

	pn.DrawScore(win)

	// Render prey
	preyPos := pn.Geom.GetVectorFromCoords(curState.Prey)
	pMat := pixel.IM
	pMat = pMat.Moved(preyPos)
	pn.PreySprite.Draw(win, pMat)

	// Render other players
	for _, player := range curState.OtherPlayers {
		pn.OtherPlayerSprite.Draw(win, pixel.IM.Moved(pn.Geom.GetVectorFromCoords(player)))
	}

	// Render player
	playerPos := pn.Geom.GetVectorFromCoords(curState.PlayerLoc)
	mat := pixel.IM
	mat = mat.Moved(playerPos)
	pn.PlayerSprite.Draw(win, mat)

}

// Sends a move as inputted by the player to the logic node
func (pn * PixelNode) SendMove (move string) {
	pn.Sender.Write([]byte(move))
}

// Listens for new game states from pixel node
func (pn * PixelNode) RunRemoteNodeListener() {
	// takes a Listener client
	// runs the Listener in a infinite loop
	node := pn.Sender

	i := 0
	var playerPos shared.GameRenderState
	for {
		i++
		buf := make([]byte, 1024)
		rlen, err := node.Read(buf)
		if err != nil {
			fmt.Println(err)
		}
		err = json.Unmarshal(buf[0:rlen], &playerPos)
		if err != nil {
			fmt.Println("Error receiving new GameRenderState on PixelNode:", err)

		} else {
			pn.NewGameStates <- playerPos
		}
	}
}

// Helper function to draw the scores on the board.
func (pn * PixelNode) DrawScore (window *pixelgl.Window) {
	pn.ScoreboardBg.Draw(window)

	fakeScoreMap := make(map[string]int)
	fakeScoreMap["1"] = 100
	fakeScoreMap["2"] = 300
	fakeScoreMap["3"] = 50

	const textHeight = 10
	const titleMultiplier = 3
	const scoreMultiplier = 1.5
	const padding = 10

	// Render the title
	titlePos := pixel.V(pn.Geom.GetX() + padding*4, pn.Geom.GetY() - titleMultiplier * textHeight)
	title := text.New(titlePos, pn.TextAtlas)
	fmt.Fprintln(title, "SCORES")
	title.Draw(window, pixel.IM.Scaled(title.Orig, titleMultiplier))

	// Render the scores
	scoreString := SortScores(fakeScoreMap) // sort 'em
	scoresPos := pixel.V(pn.Geom.GetX() + padding, pn.Geom.GetY() - (titleMultiplier + 2) * textHeight)
	scores := text.New(scoresPos, pn.TextAtlas)
	fmt.Fprintln(scores, scoreString)
	scores.Draw(window, pixel.IM)

	// Render my score
	myScoreString := fmt.Sprintf("SCORE: %10d", fakeScoreMap["2"])
	myScorePos := pixel.V(pn.Geom.GetX() + padding, textHeight * scoreMultiplier)
	myScore := text.New(myScorePos, pn.TextAtlas)
	fmt.Fprintln(myScore, myScoreString)
	myScore.Draw(window, pixel.IM.Scaled(myScore.Orig, scoreMultiplier))
}

// Helper function to take the score map and return a sorted list of all scores by player, formatted as a single string
// for pixel to draw.
// Couldn't be bothered to figure the sorting out myself, reference:
// https://stackoverflow.com/questions/18695346/how-to-sort-a-mapstringint-by-its-values
func SortScores (scoreMap map[string]int) (string) {
	n := map[int][]string{}
	var a []int
	for k, v := range scoreMap {
		n[v] = append(n[v], k)
	}
	for k := range n {
		a = append(a, k)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(a)))

	i := 1
	scoreString := ""
	for _, k := range a {
		for _, s := range n[k] {
			scoreString += fmt.Sprintf("%2d. %-4s %9d points\n\n", i, s, k)
		}
		i++
	}

	return scoreString
}

// Helper function to draw all the walls on each render update
func (pn * PixelNode ) DrawWalls(window *pixelgl.Window) {
	for _, wall := range pn.Geom.GetWallVectors() {
		pn.WallSprite.Draw(window, pixel.IM.Moved(wall))
	}
}

// Helper function to create he scoreboard background based on the initial game settings
func createScoreboard(gameWidth, gameHeight, scoreboardWidth float64) (*imdraw.IMDraw) {
	// Create scoreboard background
	scoreboardBg := imdraw.New(nil)

	scoreboardBg.Color = pixel.RGB(0, 0, 0)
	scoreboardBg.Push(pixel.V(gameWidth, 0))
	scoreboardBg.Push(pixel.V(gameWidth, gameHeight))
	scoreboardBg.Push(pixel.V(gameWidth + scoreboardWidth, gameHeight))
	scoreboardBg.Push(pixel.V(gameWidth + scoreboardWidth, 0))
	scoreboardBg.Polygon(0)
	return scoreboardBg
}

// Function that sets up the TCP connection with the logic node
func setupTCP(ip_addr string) (*net.TCPConn) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", ip_addr)
	if err != nil {
		fmt.Println("Invalid TCP address provided for logic node")
		os.Exit(1)
	}

	tcpConn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		fmt.Printf("No logic node found at %s, start a logic node and try again", tcpAddr.String())
		os.Exit(1)
	}

	return tcpConn
}
