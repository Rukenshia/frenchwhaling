package api

import (
	"errors"
	"fmt"
	"log"
	"os"

	resty "github.com/go-resty/resty/v2"
)

type ApiResponse struct {
	Status string `json:"status"`
	Meta   struct {
		Count  int         `json:"count"`
		Hidden interface{} `json:"hidden"`
	} `json:"meta"`
}

type PlayerInfoResponse struct {
	ApiResponse
	Data map[string]PlayerInfo `json:"data"`
}

type PlayerInfo struct {
	AccountID     int    `json:"account_id"`
	CreatedAt     int    `json:"created_at"`
	HiddenProfile bool   `json:"hidden_profile"`
	Nickname      string `json:"nickname"`
	Private       struct {
		Gold             int `json:"gold"`
		FreeXp           int `json:"free_xp"`
		Credits          int `json:"credits"`
		PremiumExpiresAt int `json:"premium_expires_at"`
		EmptySlots       int `json:"empty_slots"`
		Slots            int `json:"slots"`
		BattleLifeTime   int `json:"battle_life_time"`
	} `json:"private"`
}

func GetPlayerInfo(accessToken, accountId string) (*PlayerInfo, error) {
	client := resty.New()

	log.Printf("GetPlayerInfo: accountId=%s", accountId)

	res, err := client.R().
		SetResult(PlayerInfoResponse{}).
		SetQueryParam("application_id", os.Getenv("APPLICATION_ID")).
		SetQueryParam("account_id", accountId).
		SetQueryParam("access_token", accessToken).
		SetQueryParam("fields", "account_id,created_at,nickname,hidden_profile,private").
		Get("https://api.worldofwarships.eu/wows/account/info/")

	if err != nil {
		log.Printf("GetPlayerInfo: error=%v response=%s", err, res.String())
		return nil, err
	}

	data, ok := res.Result().(*PlayerInfoResponse)
	if !ok {
		log.Printf("GetPlayerInfo: error=parse failed response=%s", res.String())
		return nil, errors.New("Could not parse response from Wargaming API")
	}

	if data.Status != "ok" {
		return nil, fmt.Errorf("WG API status: %v", data)
	}

	entry := data.Data[accountId]

	return &entry, nil
}
