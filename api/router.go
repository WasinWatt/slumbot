package api

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/WasinWatt/slumbot/service"
	"github.com/WasinWatt/slumbot/user"
	"github.com/line/line-bot-sdk-go/linebot"
)

// Handler is a api handler
type Handler struct {
	Client     *linebot.Client
	db         *sql.DB
	controller *service.Controller
}

// NewHandler creates new hanlder
func NewHandler(lineClient *linebot.Client, db *sql.DB, controller *service.Controller) *Handler {
	return &Handler{
		Client:     lineClient,
		db:         db,
		controller: controller,
	}
}

// MakeHandler make default handler
func (h *Handler) MakeHandler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/line", h.lineRequestHandler())
	return mux
}

func (h *Handler) lineRequestHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		events, err := h.Client.ParseRequest(req)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		for _, event := range events {
			if event.Type == linebot.EventTypeMessage {
				userID := event.Source.UserID
				groupID := event.Source.GroupID
				replyID := userID
				if groupID != "" {
					replyID = groupID
				}

				log.Println(userID)
				res, err := h.Client.GetProfile(userID).Do()
				if err != nil {
					replyMessage(h.Client, replyID, "นายๆ แอดเพื่อนเราก่อนถึงจะใช้ได้นะ")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				username := res.DisplayName

				switch message := event.Message.(type) {
				case *linebot.TextMessage:
					err := h.handleTextMessage(message, replyID, userID, username)
					if err != nil {
						log.Println(err)
						w.Header().Set("Content-type", "application/json; charset=utf-8")
						w.WriteHeader(http.StatusInternalServerError)
						replyInternalErrorMessage(h.Client, replyID, err)
						return
					}

					w.WriteHeader(200)
				}
			}
		}
	})
}

func (h *Handler) handleTextMessage(message *linebot.TextMessage, replyID string, userID string, username string) error {
	var words []string
	words = strings.SplitN(message.Text, " ", 2)

	command := strings.ToLower(words[0])
	u, err := h.controller.GetUser(userID)
	if err == sql.ErrNoRows {
		u = &user.User{
			ID:         userID,
			Name:       username,
			PenaltyNum: 0,
		}
		h.controller.CreateUser(u)
	} else if err != nil {
		return err
	}

	if command == "เปิดตี้" || command == "เปิดโครง" || command == "create" {
		if len(words) < 2 {
			replyDefaultMessage(h.Client, replyID)
			return nil
		}
		err := h.controller.CreateRoom(words[1], userID, username)
		if err != nil {
			log.Println(err)
			return err
		}

		title := "!!! โครง " + words[1] + " เปิดแล้วจ้าา !!!"
		replyMessage(h.Client, replyID, title)

		usernames, err := h.controller.GetAllUsernamesByRoomID(words[1])
		if err != nil {
			return err
		}

		reply := "รายชื่อคนที่ไปโครง " + words[1] + " ตอนนี้ ~\n"
		for i, username := range usernames {
			reply += strconv.Itoa(i+1) + ". " + username + "\n"
		}

		replyMessage(h.Client, replyID, reply)
		replySticker(h.Client, replyID, "2", "144")
		return nil
	}

	if command == "join" || command == "ไป" || command == "ไปด้วย" || command == "จอย" {
		if len(words) < 2 {
			replyDefaultMessage(h.Client, replyID)
			return nil
		}

		err = h.controller.Join(u, words[1])
		if err != nil {
			return err
		}

		title := "! " + username + " จอยตี้ " + words[1] + " แล้ววว ~"
		replyMessage(h.Client, replyID, title)

		usernames, err := h.controller.GetAllUsernamesByRoomID(words[1])
		if err != nil {
			return err
		}

		reply := "รายชื่อคนที่ไปโครง " + words[1] + " ตอนนี้ ~\n"
		for i, username := range usernames {
			reply += strconv.Itoa(i+1) + ". " + username + "\n"
		}
		replyMessage(h.Client, replyID, reply)

		return nil
	}

	if command == "leave" || command == "เทตี้" || command == "เท" {
		if len(words) < 2 {
			replyDefaultMessage(h.Client, replyID)
			return nil
		}

		penalty, err := h.controller.Leave(u, words[1])
		if err != nil {
			return err
		}

		reply := "สายเทนะเรา " + u.Name + "\n" + "มันเทไปแล้ว " + strconv.Itoa(penalty) + " ครั้ง !!"

		replyMessage(h.Client, replyID, reply)
		replySticker(h.Client, replyID, "2", "24")
		return nil

	}
	if command == "list" || command == "ดูรายชื่อ" || command == "ใครไปบ้าง" || command == "รายชื่อ" {
		if len(words) < 2 {
			replyDefaultMessage(h.Client, replyID)
			return nil
		}

		u, err := h.controller.GetUser(userID)
		if err == sql.ErrNoRows {
			u = &user.User{
				ID:         userID,
				Name:       username,
				PenaltyNum: 0,
			}
			h.controller.CreateUser(u)
		} else if err != nil {
			return err
		}

		usernames, err := h.controller.GetAllUsernamesByRoomID(words[1])
		if err != nil {
			return err
		}

		reply := "รายชื่อคนที่ไปโครง " + words[1] + " ตอนนี้ ~\n"
		for i, username := range usernames {
			reply += strconv.Itoa(i+1) + ". " + username + "\n"
		}
		replyMessage(h.Client, replyID, reply)

		return nil
	}
	if command == "สวัสดี" || command == "hello" || command == "หวัดดี" {
		replyDefaultMessage(h.Client, replyID)
	}

	if command == "ปปขาวหรือไม่" {
		replyMessage(h.Client, replyID, "โคด")
		time.Sleep(3 * time.Second)
		replyMessage(h.Client, replyID, "ดำ")
	}

	return nil
}

func replyInternalErrorMessage(client *linebot.Client, replyID string, err error) {
	message := `ระบบขัดข้อง กรุณาลองใหม่`
	if err == service.ErrDuplicateUserInRoom {
		message = `จอยตี้ซ้ำไม่ได้นะจ๊ะ จุ้บๆ ~`
	}
	if err == service.ErrRoomNotFound {
		message = `ตี้นี้ยังไม่ได้เปิดเลยนะ งงจัง`
	}
	replyMessage(client, replyID, message)
}

func replyDefaultMessage(client *linebot.Client, replyID string) {
	message := `คุณสลัมสวัสดี โจ้ว โจ้ว ตอนนี้คุณสลัมมีความสามารถดังนี้
	
	1. อยากเปิดตี้
	- พิม "เปิดตี้/เปิดโครง" วรรคแล้วตามด้วยชื่อโครงน๊ะจ๊ะ
	ex: เปิดตี้ กินแซลมอน
	2. อยากจอยตี้
	- พิม "จอย/ไป" วรรคแล้วตามด้วยชื่อโครงเหมือนเดิม
	ex: จอย กินแซลมอน
	3. ดูรายชื่อคนไป
	- พิม "รายชื่อ/ใครไปบ้าง" วรรคแล้วตามด้วยชื่อโครง
	ex: รายชื่อ กินแซลมอน
	4. เทโครง (คุณสลัมไม่แนะนำ)
	- พิม "เท" วรรคแล้วตามด้วยชื่อโครง
	ex: เท กินแซลมอน
	
	อย่าลืมแอดเพื่อนคุณสลัมก่อนใช้งาน ขอบคุณ โจ้ว โจ้ว`
	replySticker(client, replyID, "2", "520")
	replyMessage(client, replyID, message)
}

func replySticker(client *linebot.Client, replyID string, packageID string, stickerID string) {
	client.PushMessage(replyID, linebot.NewStickerMessage(packageID, stickerID)).Do()
}

func replyMessage(client *linebot.Client, replyID string, message string) {
	client.PushMessage(replyID, linebot.NewTextMessage(message)).Do()
}
