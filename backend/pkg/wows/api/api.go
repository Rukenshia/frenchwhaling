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
	Error  struct {
		Code    interface{} `json:"code"`
		Message string      `json:"message"`
		Field   string      `json:"field"`
		Value   string      `json:"value"`
	} `json:"error"`
	Meta struct {
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

type PlayerPortResponse struct {
	ApiResponse
	Data map[string]struct {
		Private struct {
			Port []int64 `json:"port"`
		} `json:"private"`
	} `json:"data"`
}

type ShipsStatisticsResponse struct {
	ApiResponse
	Data map[string][]ShipStatistics `json:"data"`
}

type ShipStatisticsPrivate struct {
	InGarage bool `json:"in_garage"`
}

type ShipStatistics struct {
	ShipID         int64                 `json:"ship_id"`
	LastBattleTime int                   `json:"last_battle_time"`
	Battles        int                   `json:"battles"`
	Private        ShipStatisticsPrivate `json:"private"`
	Pvp            struct {
		Wins    int `json:"wins"`
		Battles int `json:"battles"`
	} `json:"pvp"`
	RankSolo struct {
		Wins    int `json:"wins"`
		Battles int `json:"battles"`
	} `json:"rank_solo"`
	OperDiv struct {
		Wins    int `json:"wins"`
		Battles int `json:"battles"`
	} `json:"oper_div"`
	Pve struct {
		Wins    int `json:"wins"`
		Battles int `json:"battles"`
	} `json:"pve"`
	OperSolo struct {
		Wins    int `json:"wins"`
		Battles int `json:"battles"`
	} `json:"oper_solo"`
}

type RefreshAccessTokenResponse struct {
	ApiResponse
	Data struct {
		AccountID   int64  `json:"account_id"`
		AccessToken string `json:"access_token"`
		ExpiresAt   int64  `json:"expires_at"`
	} `json:"data"`
}

func GetPlayerInfo(realm, accessToken, accountId string) (*PlayerInfo, error) {
	client := resty.New()

	log.Printf("GetPlayerInfo: accountId=%s realm=%s", accountId, realm)

	res, err := client.R().
		SetResult(PlayerInfoResponse{}).
		SetQueryParam("application_id", os.Getenv("APPLICATION_ID")).
		SetQueryParam("account_id", accountId).
		SetQueryParam("access_token", accessToken).
		SetQueryParam("fields", "account_id,created_at,nickname,hidden_profile,private").
		Get(fmt.Sprintf("https://api.worldofwarships.%s/wows/account/info/", realm))

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

func GetPlayerPort(realm, accessToken, accountId string) ([]int64, error) {
	client := resty.New()

	log.Printf("GetPlayerPort: accountId=%s realm=%s", accountId, realm)

	res, err := client.R().
		SetResult(PlayerPortResponse{}).
		SetQueryParam("application_id", os.Getenv("APPLICATION_ID")).
		SetQueryParam("account_id", accountId).
		SetQueryParam("access_token", accessToken).
		SetQueryParam("extra", "private.port").
		SetQueryParam("fields", "private.port").
		Get(fmt.Sprintf("https://api.worldofwarships.%s/wows/account/info/", realm))

	if err != nil {
		log.Printf("GetPlayerPort: error=%v response=%s", err, res.String())
		return nil, err
	}

	data, ok := res.Result().(*PlayerPortResponse)
	if !ok {
		log.Printf("GetPlayerPort: error=parse failed response=%s", res.String())
		return nil, errors.New("Could not parse response from Wargaming API")
	}

	if data.Status != "ok" {
		return nil, fmt.Errorf("WG API status: %v", data)
	}

	return data.Data[accountId].Private.Port, nil
}

func GetPlayerShipStatistics(realm, accessToken, accountId string) (map[int64]ShipStatistics, error) {
	client := resty.New()

	log.Printf("GetPlayerShipStatistics: accountId=%s realm=%s", accountId, realm)

	res, err := client.R().
		SetResult(ShipsStatisticsResponse{}).
		SetQueryParam("application_id", os.Getenv("APPLICATION_ID")).
		SetQueryParam("account_id", accountId).
		SetQueryParam("access_token", accessToken).
		SetQueryParam("extra", "pve,oper_solo,oper_div,rank_solo").
		SetQueryParam("fields", "ship_id,last_battle_time,battles,pvp.battles,pvp.wins,pve.battles,pve.wins,oper_solo.battles,oper_solo.wins,oper_div.battles,oper_div.wins,rank_solo.battles,rank_solo.wins,private.in_garage").
		Get(fmt.Sprintf("https://api.worldofwarships.%s/wows/ships/stats/", realm))

	if err != nil {
		log.Printf("GetPlayerShipStatistics: error=%v response=%s", err, res.String())
		return nil, err
	}

	data, ok := res.Result().(*ShipsStatisticsResponse)
	if !ok {
		log.Printf("GetPlayerShipStatistics: error=parse failed response=%s", res.String())
		return nil, errors.New("Could not parse response from Wargaming API")
	}

	if data.Status != "ok" {
		return nil, fmt.Errorf("WG API status: %v", data)
	}

	entry := data.Data[accountId]

	shipStatistics := make(map[int64]ShipStatistics)

	for _, e := range entry {
		shipStatistics[e.ShipID] = e
	}

	return shipStatistics, nil
}

func RefreshAccessToken(realm, accessToken, accountId string) (*RefreshAccessTokenResponse, error) {
	client := resty.New()

	log.Printf("RefreshAccessToken: accountId=%s realm=%s", accountId, realm)

	res, err := client.R().
		SetResult(RefreshAccessTokenResponse{}).
		SetFormData(map[string]string{
			"application_id": os.Getenv("APPLICATION_ID"),
			"access_token":   accessToken,
		}).
		Post(fmt.Sprintf("https://api.worldoftanks.%s/wot/auth/prolongate/", realm))

	if err != nil {
		log.Printf("RefreshAccessToken: error=%v response=%s", err, res.String())
		return nil, err
	}

	data, ok := res.Result().(*RefreshAccessTokenResponse)
	if !ok {
		log.Printf("RefreshAccessToken: error=parse failed response=%s", res.String())
		return nil, errors.New("Could not parse response from Wargaming API")
	}

	if data.Status != "ok" {
		return nil, fmt.Errorf("WG API status: %v", data)
	}

	return data, nil
}
