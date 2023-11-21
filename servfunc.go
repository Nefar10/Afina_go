package main

import (
	"encoding/json"
	"log"
	"os"
)

func AllowChat(chatID int64, username, text string) int {
	if !fileexists(FILES_ALLOW_LIST) {
		log.Printf("File %s not exists", gDir+FILES_ALLOW_LIST)
		file, err := os.Create(gDir + FILES_ALLOW_LIST)
		if err != nil {
			log.Fatal(err)
		}
		encoder := json.NewEncoder(file)
		err = encoder.Encode(gChatsStates)
		if err != nil {
			log.Fatal(err)
		}
		SendToOwner("Пользователь "+username+" открыл диалог.\nCообщение пользователя \n```\n"+text+"\n```\nРазрешите мне общаться с этим пользователем?", ALLOW)
		defer file.Close()

	}
	return 0
}
