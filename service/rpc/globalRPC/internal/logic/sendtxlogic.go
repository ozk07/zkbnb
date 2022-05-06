/*
 * Copyright © 2021 Zecrey Protocol
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package logic

import (
	"context"
	"fmt"
	"github.com/zecrey-labs/zecrey/service/rpc/globalRPC/globalRPCProto"
	"github.com/zecrey-labs/zecrey/service/rpc/globalRPC/internal/logic/txHandler"
	"github.com/zecrey-labs/zecrey/service/rpc/globalRPC/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type SendTxLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendTxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendTxLogic {
	return &SendTxLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func packSendTxResp(
	status int64,
	msg string,
	err string,
	result *globalRPCProto.ResultSendTx,
) (res *globalRPCProto.RespSendTx) {
	res = &globalRPCProto.RespSendTx{
		Status: status,
		Msg:    msg,
		Err:    err,
		Result: result,
	}
	return res
}

func (l *SendTxLogic) SendTx(in *globalRPCProto.ReqSendTx) (resp *globalRPCProto.RespSendTx, err error) {
	var (
		txId       string
		resultResp *globalRPCProto.ResultSendTx
	)
	switch in.TxType {
	case txHandler.TxTypeUnLock:
		txId, err = l.sendUnlockTx(in.TxInfo)

		resultResp = &globalRPCProto.ResultSendTx{
			TxId: txId,
		}

		if err != nil {
			errInfo := fmt.Sprintf("[sendtxlogic.sendUnlockTx] %s", err.Error())
			logx.Error(errInfo)
			return packSendTxResp(FailStatus, FailMsg, errInfo, resultResp), err
		}
		break
	case txHandler.TxTypeTransfer:
		txId, err = l.sendTransferTx(in.TxInfo)

		resultResp = &globalRPCProto.ResultSendTx{
			TxId: txId,
		}

		if err != nil {
			errInfo := fmt.Sprintf("[sendtxlogic.sendTransferTx] %s", err.Error())
			logx.Error(errInfo)
			return packSendTxResp(FailStatus, FailMsg, errInfo, resultResp), err
		}
		break
	case txHandler.TxTypeSwap:
		txId, err = l.sendSwapTx(in.TxInfo)

		resultResp = &globalRPCProto.ResultSendTx{
			TxId: txId,
		}

		if err != nil {
			errInfo := fmt.Sprintf("[sendtxlogic.sendSwapTx] %s", err.Error())
			logx.Error(errInfo)
			return packSendTxResp(FailStatus, FailMsg, errInfo, resultResp), err
		}
		break
	case txHandler.TxTypeAddLiquidity:
		txId, err = l.sendAddLiquidityTx(in.TxInfo)
		resultResp = &globalRPCProto.ResultSendTx{
			TxId: txId,
		}

		if err != nil {
			errInfo := fmt.Sprintf("[sendtxlogic.sendAddLiquidityTx] %s", err.Error())
			logx.Error(errInfo)
			return packSendTxResp(FailStatus, FailMsg, errInfo, resultResp), err
		}
		break
	case txHandler.TxTypeRemoveLiquidity:
		txId, err = l.sendRemoveLiquidityTx(in.TxInfo)
		resultResp = &globalRPCProto.ResultSendTx{
			TxId: txId,
		}

		if err != nil {
			errInfo := fmt.Sprintf("[sendtxlogic.sendRemoveLiquidityTx] %s", err.Error())
			logx.Error(errInfo)
			return packSendTxResp(FailStatus, FailMsg, errInfo, resultResp), err
		}
		break
	case txHandler.TxTypeWithdraw:
		txId, err = l.sendWithdrawTx(in.TxInfo)
		resultResp = &globalRPCProto.ResultSendTx{
			TxId: txId,
		}

		if err != nil {
			errInfo := fmt.Sprintf("[sendtxlogic.sendWithdrawTx] %s", err.Error())
			logx.Error(errInfo)
			return packSendTxResp(FailStatus, FailMsg, errInfo, resultResp), err
		}
		break
	}
	return packSendTxResp(SuccessStatus, SuccessMsg, "", resultResp), nil
}
