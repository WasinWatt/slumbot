package api

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/WasinWatt/slumbot/cache"

	"github.com/WasinWatt/slumbot/service"
	"github.com/WasinWatt/slumbot/user"
	"github.com/line/line-bot-sdk-go/linebot"
)

// Handler is a api handler
type Handler struct {
	Client     *linebot.Client
	db         *sql.DB
	controller *service.Controller
	memcache   cache.Cacher
}

// NewHandler creates new hanlder
func NewHandler(lineClient *linebot.Client, db *sql.DB, controller *service.Controller, c cache.Cacher) *Handler {
	return &Handler{
		Client:     lineClient,
		db:         db,
		controller: controller,
		memcache:   c,
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
					// replyMessage(h.Client, replyID, "นายๆ แอดเพื่อนเราก่อนถึงจะใช้ได้นะ")
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
	if len(words) >= 2 {
		words[1] = strings.TrimSpace(words[1])
	}

	command := strings.ToLower(words[0])

	cachename, ok := h.memcache.Get(userID)
	if !ok {
		u, err := h.controller.GetUser(userID)
		if err == sql.ErrNoRows {
			u = &user.User{
				ID:         userID,
				Name:       username,
				PenaltyNum: 0,
			}
			err := h.controller.CreateUser(u)
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		h.memcache.Set(userID, username)
	} else {
		if cachename != username {
			h.controller.UpdateUsername(userID, username)
			h.memcache.Set(userID, username)
		}
	}

	if command == "เปิดตี้" || command == "เปิดโครง" || command == "create" {
		if len(words) < 2 {
			replyDefaultMessage(h.Client, replyID)
			return nil
		}

		err := h.controller.CreateRoom(words[1], userID)
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

		err := h.controller.Join(userID, words[1])
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

		penalty, err := h.controller.Leave(userID, words[1])
		if err != nil {
			return err
		}

		reply := "สายเทนะเรา " + username + "\n" + "มันเทไปแล้ว " + strconv.Itoa(penalty) + " ครั้ง !!"

		replyMessage(h.Client, replyID, reply)
		replySticker(h.Client, replyID, "2", "24")
		return nil

	}
	if command == "list" || command == "ดูรายชื่อ" || command == "ใครไปบ้าง" || command == "รายชื่อ" {
		if len(words) < 2 {
			replyDefaultMessage(h.Client, replyID)
			return nil
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

	if command == "จบตี้" || command == "ปิดตี้" {
		if len(words) < 2 {
			replyDefaultMessage(h.Client, replyID)
			return nil
		}

		err := h.controller.DeleteRoom(words[1], userID)
		if err != nil {
			return err
		}

		reply := "ตี้ " + words[1] + " ปิดแล้วว T.T\nเจอกันตี้หน้านะทุกคนนน ไว้มาเจอคุณสลัมอีกน้าา ~"
		replyMessage(h.Client, replyID, reply)
		replySticker(h.Client, replyID, "1", "114")

	}

	if command == "รายชื่อตี้" || command == "รายชื่อโครง" || command == "มีตี้ไรบ้าง" {
		rooms, err := h.controller.ListRoomIDs()
		if err != nil {
			return err
		}

		reply := "รายชื่อโครงที่เปิดอยู่ตอนนี้ ~\n"
		for i, r := range rooms {
			reply += strconv.Itoa(i+1) + ". " + r + "\n"
		}

		replyMessage(h.Client, replyID, reply)
		return nil
	}
	if command == "สวัสดี" || command == "hello" || command == "หวัดดี" {
		replyDefaultMessage(h.Client, replyID)
	}

	if command == "จิงมั้ยคุณสลัม" || command == "จิงไหมคุนสลัม" || command == "จิงมั้ยคุนสลัม" || command == "จิงไหมคุณสลัม" {
		if userID == "U488314d7ea2adc137d8d50629beb6a47" {
			replyMessage(h.Client, replyID, "พูดอีกก็ถูกอีกคุณปั่น ฉลาดจิมๆ")
		} else if userID == "U81455a6c0ae550b54ee5fe5bfa69ef3b" {
			replyMessage(h.Client, replyID, "มั่วไปเรื่อย มึงอะ")
		} else {
			replyMessage(h.Client, replyID, "ไม่บอกหรอก")
		}

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
	if err == service.ErrDuplicateRoom {
		message = `ห้ามเปิดตี้ซ้ำนะตัวเอง`
	}
	if err == service.ErrUserNotInRoom {
		message = `มึงยังไม่ได้อยู่ในตี้เลย เทสะเปะสะปะนะเรา`
	}
	if err == service.ErrUnAuthorized {
		message = `คนเปิดตี้ หรือ adminเท่านั้นที่ปิดตี้ได้นะ เป็นเด็กเป็นเล็กอย่าเพิ่งซนนะ`
	}
	replyMessage(client, replyID, message)
}

func replyDefaultMessage(client *linebot.Client, replyID string) {
	message := `คุณสลัมสวัสดี โจ้ว โจ้ว ตอนนี้คุณสลัมมีความสามารถดังนี้
	
	1. อยากเปิดตี้
	- พิม "เปิดตี้/เปิดโครง" วรรคแล้วตามด้วยชื่อตี้น๊ะจ๊ะ
	ex: เปิดตี้ กินแซลมอน

	2. อยากจอยตี้
	- พิม "จอย/ไป" วรรคแล้วตามด้วยชื่อตี้เหมือนเดิม
	ex: จอย กินแซลมอน

	3. ดูรายชื่อคนไป
	- พิม "รายชื่อ/ใครไปบ้าง" วรรคแล้วตามด้วยชื่อตี้
	ex: รายชื่อ กินแซลมอน

	4. เทโครง (คุณสลัมไม่แนะนำ)
	- พิม "เท" วรรคแล้วตามด้วยชื่อตี้
	ex: เท กินแซลมอน

	5. ดูรายชื่อตี้ที่เปิดอยู่
	- พิม "รายชื่อตี้/มีตี้ไรบ้าง"

	6. ปิดตี้
	- พิม "จบตี้/ปิดตี้" วรรคแล้วตามด้วยชื่อตี้
	ex: ปิดตี้ กินแซลมอน
	
	อย่าลืมแอดเพื่อนคุณสลัมก่อนใช้งาน ขอบคุณ โจ้ว โจ้ว`
	replySticker(client, replyID, "3", "233")
	replyMessage(client, replyID, message)
}

func replySticker(client *linebot.Client, replyID string, packageID string, stickerID string) {
	client.PushMessage(replyID, linebot.NewStickerMessage(packageID, stickerID)).Do()
}

func replyMessage(client *linebot.Client, replyID string, message string) {
	client.PushMessage(replyID, linebot.NewTextMessage(message)).Do()
}
