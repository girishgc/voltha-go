/*
 * Copyright 2018-present Open Networking Foundation

 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at

 * http://www.apache.org/licenses/LICENSE-2.0

 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package model

import (
	"github.com/opencord/voltha-go/common/log"
	"github.com/opencord/voltha-go/protos/common"
	"github.com/opencord/voltha-go/protos/openflow_13"
	"github.com/opencord/voltha-go/protos/voltha"
	"sync"
)

type ModelTestConfig struct {
	Root      *root
	Backend   *Backend
	RootProxy *Proxy
	DbPrefix  string
	DbType    string
	DbHost    string
	DbPort    int
	DbTimeout int
}

var (
	modelTestConfig = &ModelTestConfig{
		DbPrefix: "service/voltha",
		DbType:   "etcd",
		DbHost:   "localhost",
		//DbHost:    "10.106.153.44",
		DbPort:    2379,
		DbTimeout: 5,
	}

	logports = []*voltha.LogicalPort{
		{
			Id:           "123",
			DeviceId:     "logicalport-0-device-id",
			DevicePortNo: 123,
			RootPort:     false,
		},
	}
	ports = []*voltha.Port{
		{
			PortNo:     123,
			Label:      "test-port-0",
			Type:       voltha.Port_PON_OLT,
			AdminState: common.AdminState_ENABLED,
			OperStatus: common.OperStatus_ACTIVE,
			DeviceId:   "etcd_port-0-device-id",
			Peers:      []*voltha.Port_PeerPort{},
		},
	}

	stats = &openflow_13.OfpFlowStats{
		Id: 1111,
	}
	flows = &openflow_13.Flows{
		Items: []*openflow_13.OfpFlowStats{stats},
	}
	device = &voltha.Device{
		Id:         devID,
		Type:       "simulated_olt",
		Address:    &voltha.Device_HostAndPort{HostAndPort: "1.2.3.4:5555"},
		AdminState: voltha.AdminState_PREPROVISIONED,
		Flows:      flows,
		Ports:      ports,
	}

	logicalDevice = &voltha.LogicalDevice{
		Id:         devID,
		DatapathId: 0,
		Ports:      logports,
		Flows:      flows,
	}

	devID          string
	ldevID         string
	targetDevID    string
	targetLogDevID string
)

func init() {
	log.AddPackage(log.JSON, log.WarnLevel, nil)
	log.UpdateAllLoggers(log.Fields{"instanceId": "MODEL_TEST"})

	defer log.CleanUp()

	modelTestConfig.Backend = NewBackend(
		modelTestConfig.DbType,
		modelTestConfig.DbHost,
		modelTestConfig.DbPort,
		modelTestConfig.DbTimeout,
		modelTestConfig.DbPrefix,
	)

	msgClass := &voltha.Voltha{}
	root := NewRoot(msgClass, modelTestConfig.Backend)
	//root := NewRoot(msgClass, nil)

	//if modelTestConfig.Backend != nil {
	//modelTestConfig.Root = root.Load(msgClass)
	//} else {
	modelTestConfig.Root = root
	//}

	GetProfiling().Report()

	modelTestConfig.RootProxy = modelTestConfig.Root.node.CreateProxy("/", false)
}

func commonCallback(args ...interface{}) interface{} {
	log.Infof("Running common callback - arg count: %s", len(args))

	for i := 0; i < len(args); i++ {
		log.Infof("ARG %d : %+v", i, args[i])
	}

	mutex := sync.Mutex{}
	execStatus := args[1].(*bool)

	// Inform the caller that the callback was executed
	mutex.Lock()
	*execStatus = true
	mutex.Unlock()

	return nil
}

func commonCallback2(args ...interface{}) interface{} {
	log.Infof("Running common callback - arg count: %s", len(args))

	return nil
}

func commonCallbackFunc(args ...interface{}) interface{} {
	log.Infof("Running common callback - arg count: %d", len(args))

	for i := 0; i < len(args); i++ {
		log.Infof("ARG %d : %+v", i, args[i])
	}
	execStatusFunc := args[1].(func(bool))

	// Inform the caller that the callback was executed
	execStatusFunc(true)

	return nil
}

func firstCallback(args ...interface{}) interface{} {
	name := args[0]
	id := args[1]
	log.Infof("Running first callback - name: %s, id: %s\n", name, id)
	return nil
}

func secondCallback(args ...interface{}) interface{} {
	name := args[0].(map[string]string)
	id := args[1]
	log.Infof("Running second callback - name: %s, id: %f\n", name["name"], id)
	// FIXME: the panic call seem to interfere with the logging mechanism
	//panic("Generating a panic in second callback")
	return nil
}

func thirdCallback(args ...interface{}) interface{} {
	name := args[0]
	id := args[1].(*voltha.Device)
	log.Infof("Running third callback - name: %+v, id: %s\n", name, id.Id)
	return nil
}
