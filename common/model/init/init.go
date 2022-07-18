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
 *
 */

package main

import (
	"log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/zecrey-labs/zecrey-legend/common/sysconfigName"
	"github.com/zecrey-labs/zecrey-legend/common/tree"
	"github.com/zeromicro/go-zero/core/stores/redis"

	"github.com/zecrey-labs/zecrey-legend/common/model/account"
	asset "github.com/zecrey-labs/zecrey-legend/common/model/assetInfo"
	"github.com/zecrey-labs/zecrey-legend/common/model/basic"
	"github.com/zecrey-labs/zecrey-legend/common/model/block"
	"github.com/zecrey-labs/zecrey-legend/common/model/blockForCommit"
	"github.com/zecrey-labs/zecrey-legend/common/model/blockForProof"
	"github.com/zecrey-labs/zecrey-legend/common/model/l1BlockMonitor"
	"github.com/zecrey-labs/zecrey-legend/common/model/l1TxSender"
	"github.com/zecrey-labs/zecrey-legend/common/model/l2BlockEventMonitor"
	"github.com/zecrey-labs/zecrey-legend/common/model/l2TxEventMonitor"
	"github.com/zecrey-labs/zecrey-legend/common/model/liquidity"
	"github.com/zecrey-labs/zecrey-legend/common/model/mempool"
	"github.com/zecrey-labs/zecrey-legend/common/model/nft"
	"github.com/zecrey-labs/zecrey-legend/common/model/proofSender"
	"github.com/zecrey-labs/zecrey-legend/common/model/sysconfig"
	"github.com/zecrey-labs/zecrey-legend/common/model/tx"
)

const (
	// network rpc
	BSC_Test_Network_RPC   = "http://tf-dex-preview-validator-nlb-6fd109ac8b9d390a.elb.ap-northeast-1.amazonaws.com:8545"
	Local_Test_Network_RPC = "http://127.0.0.1:8545/"

	zecreyLegendProxy = "0x4a5B8869a5A27Cf2d481E4EbeF1343f4E6C19F52" // ZecreyLegendContractAddr
	governance        = "0x6f73ee7d600285b444E6504416AB9b0B143c1Fe4" // GovernanceContractAddr
)

func initSysConfig() []*sysconfig.Sysconfig {
	return []*sysconfig.Sysconfig{
		{
			Name:      sysconfigName.SysGasFee,
			Value:     "100000000000000",
			ValueType: "string",
			Comment:   "based on BNB",
		},
		{
			Name:      sysconfigName.TreasuryAccountIndex,
			Value:     "0",
			ValueType: "int",
			Comment:   "treasury index",
		},
		{
			Name:      sysconfigName.GasAccountIndex,
			Value:     "1",
			ValueType: "int",
			Comment:   "gas index",
		},
		{
			Name:      sysconfigName.ZecreyLegendContract,
			Value:     zecreyLegendProxy,
			ValueType: "string",
			Comment:   "Zecrey contract on BSC",
		},
		// Governance Contract
		{
			Name:      sysconfigName.GovernanceContract,
			Value:     governance,
			ValueType: "string",
			Comment:   "Governance contract on BSC",
		},

		// Asset_Governance Contract
		//{
		//	Name:      sysconfigName.AssetGovernanceContract,
		//	Value:     AssetGovernanceContractAddr,
		//	ValueType: "string",
		//	Comment:   "Asset_Governance contract on BSC",
		//},

		// Verifier Contract
		//{
		//	Name:      sysconfigName.VerifierContract,
		//	Value:     VerifierContractAddr,
		//	ValueType: "string",
		//	Comment:   "Verifier contract on BSC",
		//},
		// network rpc
		{
			Name:      sysconfigName.BscTestNetworkRpc,
			Value:     BSC_Test_Network_RPC,
			ValueType: "string",
			Comment:   "BSC network rpc",
		},
		// TODO
		{
			Name:      sysconfigName.LocalTestNetworkRpc,
			Value:     Local_Test_Network_RPC,
			ValueType: "string",
			Comment:   "Local network rpc",
		},
	}
}

func initAssetsInfo() []*asset.AssetInfo {
	return []*asset.AssetInfo{
		{
			AssetId:     0,
			L1Address:   "0x00",
			AssetName:   "BNB",
			AssetSymbol: "BNB",
			Decimals:    18,
			Status:      0,
			IsGasAsset:  asset.IsGasAsset,
		},
		//{
		//	AssetId:     1,
		//	AssetName:   "LEG",
		//	AssetSymbol: "LEG",
		//	Decimals:    18,
		//	Status:      0,
		//},
		//{
		//	AssetId:     2,
		//	AssetName:   "REY",
		//	AssetSymbol: "REY",
		//	Decimals:    18,
		//	Status:      0,
		//},
	}
}

func WithRedis(redisType string, redisPass string) redis.Option {
	return func(p *redis.Redis) {
		p.Type = redisType
		p.Pass = redisPass
	}
}

var (
	redisConn      = redis.New(basic.CacheConf[0].Host, WithRedis(basic.CacheConf[0].Type, basic.CacheConf[0].Pass))
	sysconfigModel = sysconfig.NewSysconfigModel(basic.Connection, basic.CacheConf, basic.DB)
	//priceModel = price.NewPriceModel(basic.Connection, basic.CacheConf, basic.DB)
	accountModel             = account.NewAccountModel(basic.Connection, basic.CacheConf, basic.DB)
	accountHistoryModel      = account.NewAccountHistoryModel(basic.Connection, basic.CacheConf, basic.DB)
	assetInfoModel           = asset.NewAssetInfoModel(basic.Connection, basic.CacheConf, basic.DB)
	mempoolDetailModel       = mempool.NewMempoolDetailModel(basic.Connection, basic.CacheConf, basic.DB)
	mempoolModel             = mempool.NewMempoolModel(basic.Connection, basic.CacheConf, basic.DB)
	failTxModel              = tx.NewFailTxModel(basic.Connection, basic.CacheConf, basic.DB)
	txDetailModel            = tx.NewTxDetailModel(basic.Connection, basic.CacheConf, basic.DB)
	txModel                  = tx.NewTxModel(basic.Connection, basic.CacheConf, basic.DB, redisConn)
	blockModel               = block.NewBlockModel(basic.Connection, basic.CacheConf, basic.DB, redisConn)
	blockForCommitModel      = blockForCommit.NewBlockForCommitModel(basic.Connection, basic.CacheConf, basic.DB)
	blockForProofModel       = blockForProof.NewBlockForProofModel(basic.Connection, basic.CacheConf, basic.DB)
	proofSenderModel         = proofSender.NewProofSenderModel(basic.DB)
	l1BlockMonitorModel      = l1BlockMonitor.NewL1BlockMonitorModel(basic.Connection, basic.CacheConf, basic.DB)
	l2TxEventMonitorModel    = l2TxEventMonitor.NewL2TxEventMonitorModel(basic.Connection, basic.CacheConf, basic.DB)
	l2BlockEventMonitorModel = l2BlockEventMonitor.NewL2BlockEventMonitorModel(basic.Connection, basic.CacheConf, basic.DB)
	l1TxSenderModel          = l1TxSender.NewL1TxSenderModel(basic.Connection, basic.CacheConf, basic.DB)
	liquidityModel           = liquidity.NewLiquidityModel(basic.Connection, basic.CacheConf, basic.DB)
	liquidityHistoryModel    = liquidity.NewLiquidityHistoryModel(basic.Connection, basic.CacheConf, basic.DB)
	nftModel                 = nft.NewL2NftModel(basic.Connection, basic.CacheConf, basic.DB)
	offerModel               = nft.NewOfferModel(basic.Connection, basic.CacheConf, basic.DB)
	nftHistoryModel          = nft.NewL2NftHistoryModel(basic.Connection, basic.CacheConf, basic.DB)
	nftExchangeModel         = nft.NewL2NftExchangeModel(basic.Connection, basic.CacheConf, basic.DB)
	nftCollectionModel       = nft.NewL2NftCollectionModel(basic.Connection, basic.CacheConf, basic.DB)
	nftWithdrawHistoryModel  = nft.NewL2NftWithdrawHistoryModel(basic.Connection, basic.CacheConf, basic.DB)
)

func DropTables() {
	assert.Nil(nil, sysconfigModel.DropSysconfigTable())
	assert.Nil(nil, accountModel.DropAccountTable())
	assert.Nil(nil, accountHistoryModel.DropAccountHistoryTable())
	assert.Nil(nil, assetInfoModel.DropAssetInfoTable())
	assert.Nil(nil, mempoolDetailModel.DropMempoolDetailTable())
	assert.Nil(nil, mempoolModel.DropMempoolTxTable())
	assert.Nil(nil, failTxModel.DropFailTxTable())
	assert.Nil(nil, txDetailModel.DropTxDetailTable())
	assert.Nil(nil, txModel.DropTxTable())
	assert.Nil(nil, blockModel.DropBlockTable())
	assert.Nil(nil, blockForCommitModel.DropBlockForCommitTable())
	assert.Nil(nil, blockForProofModel.DropBlockForProofTable())
	assert.Nil(nil, proofSenderModel.DropProofSenderTable())
	assert.Nil(nil, l1BlockMonitorModel.DropL1BlockMonitorTable())
	assert.Nil(nil, l2TxEventMonitorModel.DropL2TxEventMonitorTable())
	assert.Nil(nil, l2BlockEventMonitorModel.DropL2BlockEventMonitorTable())
	assert.Nil(nil, l1TxSenderModel.DropL1TxSenderTable())
	assert.Nil(nil, liquidityModel.DropLiquidityTable())
	assert.Nil(nil, liquidityHistoryModel.DropLiquidityHistoryTable())
	assert.Nil(nil, nftModel.DropL2NftTable())
	assert.Nil(nil, offerModel.DropOfferTable())
	assert.Nil(nil, nftHistoryModel.DropL2NftHistoryTable())
	assert.Nil(nil, nftExchangeModel.DropL2NftExchangeTable())
	assert.Nil(nil, nftCollectionModel.DropL2NftCollectionTable())
	assert.Nil(nil, nftWithdrawHistoryModel.DropL2NftWithdrawHistoryTable())
}

func InitTable() {
	assert.Nil(nil, sysconfigModel.CreateSysconfigTable())
	assert.Nil(nil, accountModel.CreateAccountTable())
	assert.Nil(nil, accountHistoryModel.CreateAccountHistoryTable())
	assert.Nil(nil, assetInfoModel.CreateAssetInfoTable())
	assert.Nil(nil, mempoolModel.CreateMempoolTxTable())
	assert.Nil(nil, mempoolDetailModel.CreateMempoolDetailTable())
	assert.Nil(nil, failTxModel.CreateFailTxTable())
	assert.Nil(nil, blockModel.CreateBlockTable())
	assert.Nil(nil, txModel.CreateTxTable())
	assert.Nil(nil, txDetailModel.CreateTxDetailTable())
	assert.Nil(nil, blockForCommitModel.CreateBlockForCommitTable())
	assert.Nil(nil, blockForProofModel.CreateBlockForProofTable())
	assert.Nil(nil, proofSenderModel.CreateProofSenderTable())
	assert.Nil(nil, l1BlockMonitorModel.CreateL1BlockMonitorTable())
	assert.Nil(nil, l2TxEventMonitorModel.CreateL2TxEventMonitorTable())
	assert.Nil(nil, l2BlockEventMonitorModel.CreateL2BlockEventMonitorTable())
	assert.Nil(nil, l1TxSenderModel.CreateL1TxSenderTable())
	assert.Nil(nil, liquidityModel.CreateLiquidityTable())
	assert.Nil(nil, liquidityHistoryModel.CreateLiquidityHistoryTable())
	assert.Nil(nil, nftModel.CreateL2NftTable())
	assert.Nil(nil, offerModel.CreateOfferTable())
	assert.Nil(nil, nftHistoryModel.CreateL2NftHistoryTable())
	assert.Nil(nil, nftExchangeModel.CreateL2NftExchangeTable())
	assert.Nil(nil, nftCollectionModel.CreateL2NftCollectionTable())
	assert.Nil(nil, nftWithdrawHistoryModel.CreateL2NftWithdrawHistoryTable())
	rowsAffected, err := assetInfoModel.CreateAssetsInfoInBatches(initAssetsInfo())
	if err != nil {
		log.Fatal(err)
	}
	log.Println("l2 assets info rows affected:", rowsAffected)
	rowsAffected, err = sysconfigModel.CreateSysconfigInBatches(initSysConfig())
	if err != nil {
		log.Fatal(err)
	}
	log.Println("sys config rows affected:", rowsAffected)
	err = blockModel.CreateGenesisBlock(&block.Block{
		BlockCommitment:              "0000000000000000000000000000000000000000000000000000000000000000",
		BlockHeight:                  0,
		StateRoot:                    common.Bytes2Hex(tree.NilStateRoot),
		PriorityOperations:           0,
		PendingOnChainOperationsHash: "c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470",
		CommittedTxHash:              "",
		CommittedAt:                  0,
		VerifiedTxHash:               "",
		VerifiedAt:                   0,
		BlockStatus:                  block.StatusVerifiedAndExecuted,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	DropTables()
	InitTable()
}
