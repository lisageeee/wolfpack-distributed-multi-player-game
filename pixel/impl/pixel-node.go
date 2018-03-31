package impl

import (
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
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
)

var NodeAddr string // must store as global to get it into run function
var MyAddr string

// Sprite size
const spriteStep = 30

type PixelNode struct {
	//Listener          *net.UDPConn
	Sender            *net.TCPConn
	playerPosition    shared.Coord
	Geom              geometry.PixelManager
	GameState         shared.GameRenderState
	NewGameStates     chan shared.GameRenderState
	PlayerSprite      *pixel.Sprite
	WallSprite        *pixel.Sprite
	OtherPlayerSprite *pixel.Sprite
	PreySprite        *pixel.Sprite
	ScoreboardBg 	  *imdraw.IMDraw
	TextAtlas  		  *text.Atlas
}

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
func (pn * PixelNode) RenderNewState (win * pixelgl.Window) {
	curState := pn.GameState

	// Clear current render
	win.Clear(colornames.Skyblue)

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

// Couldn't be bothered to figure this out myself
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

func (pn * PixelNode ) DrawWalls(window *pixelgl.Window) {
	for _, wall := range pn.Geom.GetWallVectors() {
		pn.WallSprite.Draw(window, pixel.IM.Moved(wall))
	}
}

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
