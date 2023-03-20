package smGPAUtils

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/runtime/ioutils"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/state"
)

type blockWAL struct {
	*logger.WrappedLogger

	dir             string
	blockWALMetrics metrics.IBlockWALMetric
}

const constFileSuffix = ".blk"

func NewBlockWAL(log *logger.Logger, baseDir string, chainID isc.ChainID, blockWALMetrics metrics.IBlockWALMetric) (BlockWAL, error) {
	dir := filepath.Join(baseDir, chainID.String())
	if err := ioutils.CreateDirectory(dir, 0o777); err != nil {
		return nil, fmt.Errorf("BlockWAL cannot create folder %v: %w", dir, err)
	}

	result := &blockWAL{
		WrappedLogger:   logger.NewWrappedLogger(log),
		dir:             dir,
		blockWALMetrics: blockWALMetrics,
	}
	result.LogDebugf("BlockWAL created in folder %v", dir)
	return result, nil
}

// Overwrites, if block is already in WAL
func (bwT *blockWAL) Write(block state.Block) error {
	commitment := block.L1Commitment()
	fileName := fileName(commitment.BlockHash())
	filePath := filepath.Join(bwT.dir, fileName)
	bwT.LogDebugf("Writing block %s to wal; file name - %s", commitment, fileName)
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o666)
	if err != nil {
		bwT.blockWALMetrics.IncFailedWrites()
		return fmt.Errorf("openning file %s for writing failed: %w", fileName, err)
	}
	defer f.Close()
	blockBytes := block.Bytes()
	n, err := f.Write(blockBytes)
	if err != nil {
		bwT.blockWALMetrics.IncFailedReads()
		return fmt.Errorf("writing block data to file %s failed: %w", fileName, err)
	}
	if len(blockBytes) != n {
		bwT.blockWALMetrics.IncFailedReads()
		return fmt.Errorf("only %v of total %v bytes of block were written to file %s", n, len(blockBytes), fileName)
	}
	bwT.blockWALMetrics.IncSegments()
	return nil
}

func (bwT *blockWAL) Contains(blockHash state.BlockHash) bool {
	_, err := os.Stat(filepath.Join(bwT.dir, fileName(blockHash)))
	return err == nil
}

func (bwT *blockWAL) Read(blockHash state.BlockHash) (state.Block, error) {
	fileName := fileName(blockHash)
	filePath := filepath.Join(bwT.dir, fileName)
	f, err := os.OpenFile(filePath, os.O_RDONLY, 0o666)
	if err != nil {
		bwT.blockWALMetrics.IncFailedReads()
		return nil, fmt.Errorf("opening file %s for reading failed: %w", fileName, err)
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		bwT.blockWALMetrics.IncFailedReads()
		return nil, fmt.Errorf("reading file %s information failed: %w", fileName, err)
	}
	blockBytes := make([]byte, stat.Size())
	n, err := bufio.NewReader(f).Read(blockBytes)
	if err != nil {
		bwT.blockWALMetrics.IncFailedReads()
		return nil, fmt.Errorf("reading file %s failed: %w", fileName, err)
	}
	if int64(n) != stat.Size() {
		bwT.blockWALMetrics.IncFailedReads()
		return nil, fmt.Errorf("only %v of total %v bytes of file %s were read", n, stat.Size(), fileName)
	}
	block, err := state.BlockFromBytes(blockBytes)
	if err != nil {
		bwT.blockWALMetrics.IncFailedReads()
		return nil, fmt.Errorf("error parsing block from bytes read from file %s: %w", fileName, err)
	}
	return block, nil
}

func fileName(blockHash state.BlockHash) string {
	return blockHash.String() + constFileSuffix
}
