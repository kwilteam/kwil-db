package arweave

// this is copied from https://github.com/everFinance/goar/blob/main/utils
// I have made some manual changes to make it work with our codebase

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/everFinance/goar/types"
)

func generateItemMetaBinary(d *types.BundleItem) ([]byte, error) {
	if len(d.Signature) == 0 {
		return nil, errors.New("must be sign")
	}

	var err error
	targetBytes := []byte{}
	if d.Target != "" {
		targetBytes, err = base64Decode(d.Target)
		if err != nil {
			return nil, err
		}
		if len(targetBytes) != 32 {
			return nil, errors.New("targetBytes length must 32")
		}
	}
	anchorBytes := []byte{}
	if d.Anchor != "" {
		anchorBytes, err = base64Decode(d.Anchor)
		if err != nil {
			return nil, err
		}
		if len(anchorBytes) != 32 {
			return nil, errors.New("anchorBytes length must 32")
		}
	}
	tagsBytes := make([]byte, 0)
	if len(d.Tags) > 0 {
		tagsBytes, err = base64Decode(d.TagsBy)
		if err != nil {
			return nil, err
		}
	}

	sigMeta, ok := types.SigConfigMap[d.SignatureType]
	if !ok {
		return nil, fmt.Errorf("not support sigType:%d", d.SignatureType)
	}

	sigLength := sigMeta.SigLength
	ownerLength := sigMeta.PubLength

	// Create array with set length
	bytesArr := make([]byte, 0, 2+sigLength+ownerLength)

	bytesArr = append(bytesArr, shortTo2ByteArray(d.SignatureType)...)
	// Push bytes for `signature`
	sig, err := base64Decode(d.Signature)
	if err != nil {
		return nil, err
	}

	if len(sig) != sigLength {
		return nil, errors.New("signature length incorrect")
	}

	bytesArr = append(bytesArr, sig...)
	// Push bytes for `ownerByte`
	ownerByte, err := base64Decode(d.Owner)
	if err != nil {
		return nil, err
	}
	if len(ownerByte) != ownerLength {
		return nil, errors.New("signature length incorrect")
	}
	bytesArr = append(bytesArr, ownerByte...)
	// Push `presence byte` and push `target` if present
	// 64 + OWNER_LENGTH
	if d.Target != "" {
		bytesArr = append(bytesArr, byte(1))
		bytesArr = append(bytesArr, targetBytes...)
	} else {
		bytesArr = append(bytesArr, byte(0))
	}

	// Push `presence byte` and push `anchor` if present
	// 64 + OWNER_LENGTH
	if d.Anchor != "" {
		bytesArr = append(bytesArr, byte(1))
		bytesArr = append(bytesArr, anchorBytes...)
	} else {
		bytesArr = append(bytesArr, byte(0))
	}

	// push tags
	bytesArr = append(bytesArr, longTo8ByteArray(len(d.Tags))...)
	bytesArr = append(bytesArr, longTo8ByteArray(len(tagsBytes))...)

	if len(d.Tags) > 0 {
		bytesArr = append(bytesArr, tagsBytes...)
	}
	return bytesArr, nil
}

func longTo8ByteArray(long int) []byte {
	// we want to represent the input as a 8-bytes array
	byteArray := []byte{0, 0, 0, 0, 0, 0, 0, 0}
	for i := 0; i < len(byteArray); i++ {
		byt := long & 0xff
		byteArray[i] = byte(byt)
		long = (long - byt) / 256
	}
	return byteArray
}

func shortTo2ByteArray(long int) []byte {
	byteArray := []byte{0, 0}
	for i := 0; i < len(byteArray); i++ {
		byt := long & 0xff
		byteArray[i] = byte(byt)
		long = (long - byt) / 256
	}
	return byteArray
}

func submitItemToBundlr(item types.BundleItem, bundlrUrl string, currencyTicker string) (*types.BundlrResp, error) {
	itemBinary := item.ItemBinary
	if len(itemBinary) == 0 {
		var err error
		itemBinary, err = generateItemBinary(&item)
		if err != nil {
			return nil, err
		}
	}
	// post to bundler
	resp, err := http.DefaultClient.Post(bundlrUrl+"/tx/"+currencyTicker, "application/octet-stream", bytes.NewReader(itemBinary))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("send to bundler request failed; http code: %d", resp.StatusCode)
	}

	defer resp.Body.Close()
	// json unmarshal
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ioutil.ReadAll(resp.Body) error: %v", err)
	}
	br := &types.BundlrResp{}
	if err := json.Unmarshal(body, br); err != nil {
		return nil, fmt.Errorf("json.Unmarshal(body,br) failed; err: %v", err)
	}
	return br, nil
}

func generateItemBinary(d *types.BundleItem) (by []byte, err error) {
	metaBinary, err := generateItemMetaBinary(d)
	if err != nil {
		return nil, err
	}

	by = append(by, metaBinary...)
	// push data
	data := make([]byte, 0)
	if len(d.Data) > 0 {
		data, err = base64Decode(d.Data)
		if err != nil {
			return nil, err
		}
		by = append(by, data...)
	}
	return
}

func base64Decode(data string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(data)
}
