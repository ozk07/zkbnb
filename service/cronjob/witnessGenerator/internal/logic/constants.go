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
	"time"

	cryptoBlock "github.com/zecrey-labs/zecrey-crypto/zecrey-legend/circuit/bn254/block"
	zecreyLegend "github.com/zecrey-labs/zecrey-eth-rpc/zecrey/core/zecrey-legend"

	"github.com/zecrey-labs/zecrey-legend/common/model/block"
	"github.com/zecrey-labs/zecrey-legend/common/model/blockForProof"
)

const (
	UnprovedBlockReceivedTimeout = 10 * time.Minute

	BlockProcessDelta = 10
)

type (
	Block                  = block.Block
	CryptoTx               = cryptoBlock.Tx
	CryptoBlock            = cryptoBlock.Block
	BlockForProof          = blockForProof.BlockForProof
	StorageStoredBlockInfo = zecreyLegend.StorageStoredBlockInfo
)

type CryptoBlockInfo struct {
	BlockInfo *CryptoBlock
	Status    int64
}