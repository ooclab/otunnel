package udp

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"math/rand"
	"testing"
)

func Test_msgSending_IterBufferd(t *testing.T) {
	for length := 2; length <= maxMsgSize; length *= 2 {
		b := make([]byte, length+1)
		rand.Read(b)
		sending := newMsgSending(0, 0, 0, 0, b)
		var buf bytes.Buffer
		for seg := range sending.IterBufferd() {
			if seg.h.OrderID() == 0 {
				buf.Write(seg.b[4:])
			} else {
				buf.Write(seg.b)
			}
		}
		if !bytes.Equal(b, buf.Bytes()) {
			t.Errorf("IterBufferd mismatch")
		}
	}
}

func Test_msgSending_GetSegmentByOrderID(t *testing.T) {
	for length := 2; length <= maxMsgSize; length *= 2 {
		b := make([]byte, length+1)
		rand.Read(b)

		sending := newMsgSending(0, 0, 0, 0, b)
		sl := map[uint16]string{}
		for seg := range sending.IterBufferd() {
			ck := md5.Sum(seg.b)
			sl[seg.h.OrderID()] = hex.EncodeToString(ck[:])
		}
		for orderID, origCk := range sl {
			seg := sending.GetSegmentByOrderID(orderID)
			ck := md5.Sum(seg.b)
			newCk := hex.EncodeToString(ck[:])
			if newCk != origCk {
				t.Errorf("GetSegmentByOrderID mismatch: orderID = %d, orig ck = %s, new ck = %s", orderID, origCk, newCk)
			}
		}
	}
}

func Test_msgRecving_Save(t *testing.T) {
	for length := 2; length <= maxMsgSize; length *= 2 {
		b := make([]byte, length+1)
		rand.Read(b)

		sending := newMsgSending(0, 0, 0, 0, b)
		sl := map[uint16]*segment{} // map for random select
		for seg := range sending.IterBufferd() {
			sl[seg.h.OrderID()] = seg
		}

		recving := newMsgRecving()
		func() {
			for _, seg := range sl {
				msg, err := recving.Save(seg)
				if err != nil {
					t.Errorf("Save failed: %s", err)
				}
				if msg != nil {
					if !bytes.Equal(msg, b) {
						t.Errorf("recving msg mismatch: length = %d", length)
					}
					return // break inner for
				}
			}
			t.Errorf("recving not completed!")
		}()
	}
}

func Test_msgRecving_GetMissing(t *testing.T) {
	for length := 2; length <= maxMsgSize; length *= 2 {
		b := make([]byte, length+1)
		rand.Read(b)

		sending := newMsgSending(0, 0, 0, 0, b)
		sl := map[uint16]*segment{} // map for random select
		for seg := range sending.IterBufferd() {
			sl[seg.h.OrderID()] = seg
		}

		recving := newMsgRecving()

		// unradom save part segments
		maxOrderID := sending.segmentCount() - 1
		ul := []uint16{} // unradom orderID list
		for i := uint16(0); i < (maxOrderID+1)/2; {
			orderID := uint16(rand.Intn(int(maxOrderID)))
			exist := func() bool {
				for _, v := range ul {
					if v == orderID {
						return true
					}
				}
				return false
			}()
			if exist {
				continue
			}
			ul = append(ul, orderID)
			i++
		}

		for _, orderID := range ul {
			seg := sending.GetSegmentByOrderID(orderID)
			_, err := recving.Save(seg)
			if err != nil {
				t.Errorf("recving save failed: %s", err)
			}
		}

		// save remains
		largestOrderID, missingOrderList := recving.GetMissing()
		for orderID := largestOrderID + 1; orderID <= maxOrderID; orderID++ {
			existed := func() bool {
				for _, v := range missingOrderList {
					if v == orderID {
						return true
					}
				}
				return false
			}()
			if !existed {
				missingOrderList = append(missingOrderList, orderID)
			}
		}
		// FIXME!
		if largestOrderID == 0 {
			missingOrderList = append(missingOrderList, 0)
		}
		func() {
			for _, orderID := range missingOrderList {
				seg := sending.GetSegmentByOrderID(orderID)
				msg, err := recving.Save(seg)
				if err != nil {
					t.Errorf("Save failed: %s", err)
				}
				if msg != nil {
					if !bytes.Equal(msg, b) {
						t.Errorf("recving msg mismatch: length = %d", length)
					}
					return // break inner for
				}
			}
			t.Errorf("recving not completed!")
		}()
	}
}
