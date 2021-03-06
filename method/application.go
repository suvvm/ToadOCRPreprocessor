package method

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"strconv"
	"suvvm.work/ToadOCRPreprocessor/dal/cluster"
	"suvvm.work/ToadOCRPreprocessor/dal/db"
	"suvvm.work/ToadOCRPreprocessor/model"
)

func VerifySecret(ctx context.Context, appID, basicToken, verifyStr string) error {
	idStr, err := strconv.Atoi(appID)
	if err != nil {
		return fmt.Errorf("appID not int %v", err)
	}
	appInfo := &model.AppInfo{}
	appInfo.ID = idStr
	appSecret, err := cluster.GetKV(ctx, strconv.Itoa(appInfo.ID))
	if err != nil {
		appInfo, err = db.GetAppInfo(appInfo)
		if err != nil {
			return fmt.Errorf("appID not exists %v", err)
		}
		cluster.PutKV(ctx, strconv.Itoa(appInfo.ID), appSecret)
		appSecret = appInfo.Secret
	}
	hasher := md5.New()
	hasher.Write([]byte(appSecret + verifyStr))
	md5Token := hex.EncodeToString(hasher.Sum(nil))
	log.Printf("VerifySecret md5Token:%v", md5Token)
	if  md5Token != basicToken {
		return fmt.Errorf("basic token incompatible")
	}
	return nil
}

