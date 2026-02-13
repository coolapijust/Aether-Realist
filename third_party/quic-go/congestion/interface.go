package congestion

import (
	"time"

	"github.com/quic-go/quic-go/internal/monotime"
	"github.com/quic-go/quic-go/internal/protocol"
)

type ByteCount = protocol.ByteCount
type Time = monotime.Time
type PacketNumber = protocol.PacketNumber

const InitialPacketSize = protocol.InitialPacketSize
const MinPacingDelay = protocol.MinPacingDelay

type AckedPacketInfo struct {
	PacketNumber protocol.PacketNumber
	AckedBytes   protocol.ByteCount
	OnLastAcked  func(protocol.ByteCount)
}

type LostPacketInfo struct {
	PacketNumber protocol.PacketNumber
	LostBytes    protocol.ByteCount
}

type RTTStatsProvider interface {
	SmoothedRTT() time.Duration
	LatestRTT() time.Duration
	MeanDeviation() time.Duration
	MinRTT() time.Duration
}

// A SendAlgorithm performs congestion control
type SendAlgorithm interface {
	TimeUntilSend(bytesInFlight ByteCount) Time
	HasPacingBudget(now Time) bool
	OnPacketSent(sentTime Time, bytesInFlight ByteCount, packetNumber PacketNumber, bytes ByteCount, isRetransmittable bool)
	CanSend(bytesInFlight ByteCount) bool
	MaybeExitSlowStart()
	OnPacketAcked(number PacketNumber, ackedBytes ByteCount, priorInFlight ByteCount, eventTime Time)
	OnCongestionEvent(number PacketNumber, lostBytes ByteCount, priorInFlight ByteCount)
	OnCongestionEventEx(priorInFlight ByteCount, eventTime Time, ackedPackets []AckedPacketInfo, lostPackets []LostPacketInfo)
	OnRetransmissionTimeout(packetsRetransmitted bool)
	SetMaxDatagramSize(ByteCount)
	GetCongestionWindow() ByteCount
}

// CongestionControl is an alias for SendAlgorithm
type CongestionControl = SendAlgorithm

// A SendAlgorithmWithDebugInfos is a SendAlgorithm that exposes some debug infos
type SendAlgorithmWithDebugInfos interface {
	SendAlgorithm
	InSlowStart() bool
	InRecovery() bool
	GetCongestionWindow() ByteCount
}

