package open_falcon_sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func PostPush(L []*JsonMetaData) error {
	bs, err := json.Marshal(L)
	if err != nil {
		return err
	}

	bf := bytes.NewBuffer(bs)

	//fmt.Println(string(bs))
	resp, err := http.Post(PostPushUrl, "application/json", bf)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return err
	}

	content := string(body)

	if resp.StatusCode != 200 {
		return fmt.Errorf("status code %d != 200, response: %s", resp.StatusCode, content)
	}

	if Debug {
		log.Println("[D] response:", content)
	}

	return nil
}
