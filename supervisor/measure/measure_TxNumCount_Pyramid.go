package measure

import "blockEmulator/message"

// to test cross-transaction rate
type TestTxNumCount_Pyramid struct {
	epochID      int
	txNum        []float64
	pyramidTxNum []float64
}

func NewTestTxNumCount_Pyramid() *TestTxNumCount_Pyramid {
	return &TestTxNumCount_Pyramid{
		epochID:      -1,
		txNum:        make([]float64, 0),
		pyramidTxNum: make([]float64, 0),
	}
}

func (ttnc *TestTxNumCount_Pyramid) OutputMetricName() string {
	return "Tx_number"
}

func (ttnc *TestTxNumCount_Pyramid) UpdateMeasureRecord(b *message.BlockInfoMsg) {
	if b.BlockBodyLength == 0 { // empty block
		return
	}
	epochid := b.Epoch
	// extend
	for ttnc.epochID < epochid {
		ttnc.txNum = append(ttnc.txNum, 0)
		ttnc.pyramidTxNum = append(ttnc.pyramidTxNum, 0)
		ttnc.epochID++
	}

	ttnc.txNum[epochid] += float64(len(b.ExcutedTxs))
	ttnc.txNum[epochid] += float64(len(b.PyramidTxs) / 2)

	ttnc.pyramidTxNum[epochid] += float64(len(b.PyramidTxs))
}

func (ttnc *TestTxNumCount_Pyramid) HandleExtraMessage([]byte) {}

func (ttnc *TestTxNumCount_Pyramid) OutputRecord() (perEpochCTXs []float64, totTxNum float64) {
	perEpochCTXs = make([]float64, 0)
	totTxNum = 0.0
	for _, tn := range ttnc.txNum {
		perEpochCTXs = append(perEpochCTXs, tn)
		totTxNum += tn
	}

	for _, tn := range ttnc.pyramidTxNum {
		perEpochCTXs = append(perEpochCTXs, tn*-1.0/2)
	}
	return perEpochCTXs, totTxNum
}
