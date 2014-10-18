package beehive

type cmdStop struct{}

type cmdStart struct{}

type cmdPingHive struct{}

type cmdLiveHives struct{}

type cmdCreateHiveID struct{}

type cmdProcessRaftMessage struct {
	Msg []byte
}

type cmdFindBee struct {
	ID uint64
}

type cmdCreateBee struct{}

type cmdReloadBee struct {
	ID uint64
}

// FIXME REFACTOR
//type joinColonyCmd struct {
//Colony BeeColony
//}

type cmdStartDetached struct {
	Handler DetachedHandler
}

//type bufferTxCmd struct {
//Tx Tx
//}

//type commitTxCmd struct {
//Seq TxSeq
//}

//type getTxInfoCmd struct{}

//type getTx struct {
//From TxSeq
//To   TxSeq
//}

//type migrateBeeCmd struct {
//From BeeID
//To   HiveID
//}

//type replaceBeeCmd struct {
//OldBees     BeeColony
//NewBees     BeeColony
//State       *inMemoryState
//TxBuf       []Tx
//MappedCells MappedCells
//}

//type lockMappedCellsCmd struct {
//Colony      BeeColony
//MappedCells MappedCells
//}

//type getColonyCmd struct{}

//type addSlaveCmd struct {
//BeeID
//}

//type delSlaveCmd struct {
//BeeID
//}

//type addMappedCells struct {
//Cells MappedCells
//}

//func (h *hive) createBee(hive HiveID, app AppName) (BeeID,
//error) {

//prx, err := h.newProxy(hive)
//if err != nil {
//return BeeID{}, err
//}

//to := BeeID{
//HiveID:  hive,
//AppName: app,
//}
//cmd := NewRemoteCmd(createBeeCmd{}, to)
//d, err := prx.sendCmd(&cmd)
//if err != nil {
//return BeeID{}, err
//}

//return d.(BeeID), nil
//}