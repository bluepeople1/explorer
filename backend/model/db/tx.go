package db

import (
	"encoding/json"
	"github.com/globalsign/mgo/bson"
	"github.com/iost-official/Go-IOS-Protocol/account"
	"github.com/iost-official/Go-IOS-Protocol/common"
	"github.com/iost-official/explorer/backend/model/blkchain"
)

type ActionRaw struct {
	Contract   string `bson:"contract" json:"contract"`
	ActionName string `bson:"actionName" json:"actionName"`
	Data       string `bson:"data" json:"data"`
}

type SignatureRaw struct {
	Algorithm int32  `bson:"algorithm" json:"algorithm"`
	Sig       string `bson:"sig" json:"sig"`
	PubKey    string `bson:"pubKey" json:"pubKey"`
}

type ReceiptRaw struct {
	Type    int32  `bson:"type" json:"type"`
	Content string `bson:"content" json:"content"`
}

type TxReceiptRaw struct {
	GasUsage      int64        `bson:"gasUsage"`
	SuccActionNum int32        `bson:"succActionNum"`
	Receipts      []ReceiptRaw `bson:"receipts"`
	StatusCode    int32        `bson:"statusCode"`
	StatusMessage string       `bson:"statusMessage"`
}

type Tx struct {
	BlockNumber int64          `bson:"blockNumber"`
	Time        int64          `bson:"time"`
	Hash        string         `bson:"hash"`
	Expiration  int64          `bson:"expiration"`
	GasPrice    int64          `bson:"gasPrice"`
	GasLimit    int64          `bson:"gasLimit"`
	Actions     []ActionRaw    `bson:"actions"`
	Signers     []string       `bson:"signers"`
	Signs       []SignatureRaw `bson:"signs"`
	Publisher   SignatureRaw   `bson:"publisher"`
	Receipt     TxReceiptRaw   `bson:"receipt"`
}

// 将 Tx.Actions 打平后的数据结构， 如果actionName == Transfer 则会解析出 from, to, amount
type FlatTx struct {
	Id          bson.ObjectId  `bson:"_id,omitempty" json:"id"`
	BlockNumber int64          `bson:"blockNumber" json:"blockNumber"`
	Time        int64          `bson:"time" json:"time"`
	Hash        string         `bson:"hash" json:"hash"`
	Expiration  int64          `bson:"expiration" json:"expiration"`
	GasPrice    int64          `bson:"gasPrice" json:"gasPrice"`
	GasLimit    int64          `bson:"gasLimit" json:"gasLimit"`
	Action      ActionRaw      `bson:"action" json:"action"`
	Signers     []string       `bson:"signers" json:"signers"`
	Signs       []SignatureRaw `bson:"signs" json:"signs"`
	Publisher   string         `bson:"publisher" json:"publisher"`
	From        string         `bson:"from" json:"from"`
	To          string         `bson:"to" json:"to"`
	Amount      float64        `bson:"amount" json:"amount"`      // 转发数量
	ActionIndex int            `bson:"actionIndex" json:"actionIndex"` // action 索引
	ActionName  string         `bson:"actionName" json:"actionName"`  // action 类型
}

func RpcGetTxByHash(txHash string) (*Tx, error) {
	txRes, err := blkchain.GetTxByHash(txHash)
	if err != nil {
		return nil, err
	}
	txRaw := txRes.TxRaw
	actions := make([]ActionRaw, len(txRaw.Actions))
	for i, v := range txRaw.Actions {
		actions[i] = ActionRaw{
			Contract:   v.Contract,
			ActionName: v.ActionName,
			Data:       v.Data,
		}
	}
	publisher := SignatureRaw{
		Algorithm: txRaw.Publisher.Algorithm,
		Sig:       common.Base58Encode(txRaw.Publisher.Sig),
		PubKey:    common.Base58Encode(txRaw.Publisher.PubKey),
	}
	signs := make([]SignatureRaw, len(txRaw.Signs))
	for i, v := range txRaw.Signs {
		signs[i] = SignatureRaw{
			Algorithm: v.Algorithm,
			Sig:       common.Base58Encode(v.Sig),
			PubKey:    common.Base58Encode(v.PubKey),
		}
	}
	receiptRaw, err := blkchain.GetTxReceiptByTxHash(txHash)
	if err != nil {
		return nil, err
	}
	receiptContentRaws := make([]ReceiptRaw, len(receiptRaw.TxReceiptRaw.Receipts))
	for i, v := range receiptRaw.TxReceiptRaw.Receipts{
		receiptContentRaws[i] = ReceiptRaw{
			Type: v.Type,
			Content: v.Content,
		}
	}
	receipt := TxReceiptRaw{
		GasUsage: receiptRaw.TxReceiptRaw.GasUsage,
		SuccActionNum: receiptRaw.TxReceiptRaw.SuccActionNum,
		StatusCode: receiptRaw.TxReceiptRaw.Status.Code,
		StatusMessage: receiptRaw.TxReceiptRaw.Status.Message,
		Receipts: receiptContentRaws,
	}
	return &Tx{
		Time:       txRaw.Time,
		Hash:       txHash,
		Expiration: txRaw.Expiration,
		GasPrice:   txRaw.GasPrice,
		GasLimit:   txRaw.GasLimit,
		Actions:    actions,
		Signers:    byteSliceArrayToStringArray(txRaw.Signers),
		Signs:      signs,
		Publisher:  publisher,
		Receipt:    receipt,
	}, nil
}

func (tx *Tx) ToFlatTx() []*FlatTx {
	flatTx := make([]*FlatTx, len(tx.Actions))

	for i, v := range tx.Actions {
		var from, to string
		var amount float64
		if v.ActionName == "Transfer" {
			var tmp []interface{}
			json.Unmarshal([]byte(v.Data), &tmp) // TODO check error
			from = tmp[0].(string)
			to = tmp[1].(string)
			amount = tmp[2].(float64)
		}

		pubKey := common.Base58Decode(tx.Publisher.PubKey)
		publisher := account.GetIDByPubkey([]byte(pubKey))
		flatTx[i] = &FlatTx{
			BlockNumber: tx.BlockNumber,
			Time:        tx.Time,
			Hash:        tx.Hash,
			Expiration:  tx.Expiration,
			GasPrice:    tx.GasPrice,
			GasLimit:    tx.GasLimit,
			Signers:     tx.Signers,
			Publisher:   publisher,
			Signs:       tx.Signs,
			Action:      v,
			From:        from,
			To:          to,
			Amount:      amount,
			ActionIndex: i,
			ActionName:  v.ActionName,
		}
	}
	return flatTx
}

func byteSliceArrayToStringArray(origin [][]byte) []string {
	vsm := make([]string, len(origin))
	for i, v := range origin {
		vsm[i] = common.Base58Encode(v)
	}
	return vsm
}