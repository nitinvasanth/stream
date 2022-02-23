package trello

import (
	"errors"
	"fmt"
	"os"

	"github.com/adlio/trello"

	"github.com/merico-dev/stream/internal/pkg/backend"
	"github.com/merico-dev/stream/internal/pkg/log"
	"github.com/merico-dev/stream/internal/pkg/statemanager"
)

type Client struct {
	*trello.Client
}

func NewClient() (*Client, error) {
	helpUrl := "https://docs.servicenow.com/bundle/quebec-it-asset-management/page/product/software-asset-management2/task/generate-trello-apikey-token.html"
	apiKey := os.Getenv("TRELLO_API_KEY")
	token := os.Getenv("TRELLO_TOKEN")
	if apiKey == "" || token == "" {
		return nil, fmt.Errorf("TRELLO_API_KEY and/or TRELLO_TOKEN are/is empty. see %s for more info", helpUrl)
	}

	return &Client{
		Client: trello.NewClient(apiKey, token),
	}, nil
}

func (c *Client) CreateBoard(boardName string) (*trello.Board, error) {
	if boardName == "" {
		return nil, fmt.Errorf("board name can't be empty")
	}
	board := trello.NewBoard(boardName)
	err := c.Client.CreateBoard(&board, trello.Defaults())
	if err != nil {
		return nil, err
	}
	return &board, nil
}

func (c *Client) CreateList(board *trello.Board, listName string) (*trello.List, error) {
	if listName == "" {
		return nil, fmt.Errorf("listName name can't be empty")
	}
	return c.Client.CreateList(board, listName, trello.Defaults())
}

func (c *Client) GetBoardIdAndListId(ident, category string) (map[string]interface{}, error) {

	res := make(map[string]interface{})

	boardId, err := genBoardId(ident, category)
	if err != nil {
		return nil, err
	}

	b, err := c.Client.GetBoard(boardId)
	if err != nil {
		return nil, err
	}

	lists, err := b.GetLists()
	if err != nil {
		return nil, err
	}

	if len(lists) != 3 {
		log.Errorf("Unknown lists format: len==%d", len(lists))
		return nil, fmt.Errorf("unknown lists format: len==%d", len(lists))
	}
	res["boardId"] = b.ID
	res["todoListId"] = lists[0].ID
	res["doingListId"] = lists[1].ID
	res["doneListId"] = lists[2].ID

	return res, nil
}

func genBoardId(ident, category string) (string, error) {
	// use default local backend for now.
	b, err := backend.GetBackend(backend.BackendLocal)
	if err != nil {
		return "", err
	}
	// create a state manager using the default local backend
	smgr, err := statemanager.NewManager(b)
	if err != nil {
		log.Debugf("Failed to get the manager. %s", err)
		return "", err
	}

	state := smgr.GetState(fmt.Sprintf("%s_%s", ident, category))

	value, ok := state.Resource["boardId"]
	if ok {
		if value == "" {
			return "", errors.New("Board id in trello is empty.")
		}
		return fmt.Sprintf("%v", value), nil
	}
	return "", errors.New("Got some errors when generate board id form state resource.")
}
