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
package core

import (
	"context"
	"github.com/golang/protobuf/ptypes"
	a "github.com/golang/protobuf/ptypes/any"
	"github.com/opencord/voltha-go/common/log"
	"github.com/opencord/voltha-go/kafka"
	ic "github.com/opencord/voltha-go/protos/inter_container"
	"github.com/opencord/voltha-go/protos/openflow_13"
	"github.com/opencord/voltha-go/protos/voltha"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AdapterProxy struct {
	TestMode     bool
	kafkaICProxy *kafka.InterContainerProxy
}

func NewAdapterProxy(kafkaProxy *kafka.InterContainerProxy) *AdapterProxy {
	var proxy AdapterProxy
	proxy.kafkaICProxy = kafkaProxy
	return &proxy
}

func unPackResponse(rpc string, deviceId string, success bool, response *a.Any) error {
	if success {
		return nil
	} else {
		unpackResult := &ic.Error{}
		var err error
		if err = ptypes.UnmarshalAny(response, unpackResult); err != nil {
			log.Warnw("cannot-unmarshal-response", log.Fields{"error": err})
		}
		log.Debugw("response", log.Fields{"rpc": rpc, "deviceId": deviceId, "success": success, "error": err})
		// TODO:  Need to get the real error code
		return status.Errorf(codes.Canceled, "%s", unpackResult.Reason)
	}
}

//func kafka.CreateSubTopic(args ...string) kafka.Topic{
//	topic := ""
//	for index , arg := range args {
//		if index == 0 {
//			topic = arg
//		} else {
//			topic = fmt.Sprintf("%s_%s",  topic, arg)
//		}
//	}
//	return kafka.Topic{Name:topic}
//}

func (ap *AdapterProxy) AdoptDevice(ctx context.Context, device *voltha.Device) error {
	log.Debugw("AdoptDevice", log.Fields{"device": device})
	rpc := "adopt_device"
	topic := kafka.Topic{Name: device.Type}
	args := make([]*kafka.KVArg, 1)
	args[0] = &kafka.KVArg{
		Key:   "device",
		Value: device,
	}
	// Use a device topic for the response as we are the only core handling requests for this device
	replyToTopic := kafka.CreateSubTopic(ap.kafkaICProxy.DefaultTopic.Name, device.Id)
	if err := ap.kafkaICProxy.SubscribeWithDefaultRequestHandler(replyToTopic); err != nil {
		log.Errorw("Unable-to-subscribe-new-topic", log.Fields{"topic": replyToTopic, "error": err})
		return err
	}
	success, result := ap.kafkaICProxy.InvokeRPC(ctx, rpc, &topic, &replyToTopic, true, args...)
	log.Debugw("AdoptDevice-response", log.Fields{"replyTopic": replyToTopic, "deviceid": device.Id, "success": success})
	//if success {
	//	// From now on, any unsolicited requests from the adapters for this device will come over the device topic.
	//	// We should therefore include the replyToTopic as part of the target when unsolicited messages come in.
	//	if err := ap.kafkaICProxy.SubscribeWithDefaultRequestHandler(replyToTopic); err != nil {
	//		log.Errorw("Unable-to-subscribe-new-topic", log.Fields{"topic": replyToTopic, "error": err})
	//		return err
	//	}
	//}
	return unPackResponse(rpc, device.Id, success, result)
}

func (ap *AdapterProxy) DisableDevice(ctx context.Context, device *voltha.Device) error {
	log.Debugw("DisableDevice", log.Fields{"deviceId": device.Id})
	rpc := "disable_device"
	// Use a device specific topic to send the request.  The adapter handling the device creates a device
	// specific topic
	toTopic := kafka.CreateSubTopic(device.Type, device.Id)
	args := make([]*kafka.KVArg, 1)
	args[0] = &kafka.KVArg{
		Key:   "device",
		Value: device,
	}
	// Use a device specific topic as we are the only core handling requests for this device
	replyToTopic := kafka.CreateSubTopic(ap.kafkaICProxy.DefaultTopic.Name, device.Id)
	success, result := ap.kafkaICProxy.InvokeRPC(nil, rpc, &toTopic, &replyToTopic, true, args...)
	log.Debugw("DisableDevice-response", log.Fields{"deviceId": device.Id, "success": success})
	return unPackResponse(rpc, device.Id, success, result)
}

func (ap *AdapterProxy) ReEnableDevice(ctx context.Context, device *voltha.Device) error {
	log.Debugw("ReEnableDevice", log.Fields{"deviceId": device.Id})
	rpc := "reenable_device"
	toTopic := kafka.CreateSubTopic(device.Type, device.Id)
	args := make([]*kafka.KVArg, 1)
	args[0] = &kafka.KVArg{
		Key:   "device",
		Value: device,
	}
	// Use a device specific topic as we are the only core handling requests for this device
	replyToTopic := kafka.CreateSubTopic(ap.kafkaICProxy.DefaultTopic.Name, device.Id)
	success, result := ap.kafkaICProxy.InvokeRPC(ctx, rpc, &toTopic, &replyToTopic, true, args...)
	log.Debugw("ReEnableDevice-response", log.Fields{"deviceid": device.Id, "success": success})
	return unPackResponse(rpc, device.Id, success, result)
}

func (ap *AdapterProxy) RebootDevice(ctx context.Context, device *voltha.Device) error {
	log.Debugw("RebootDevice", log.Fields{"deviceId": device.Id})
	rpc := "reboot_device"
	toTopic := kafka.CreateSubTopic(device.Type, device.Id)
	args := make([]*kafka.KVArg, 1)
	args[0] = &kafka.KVArg{
		Key:   "device",
		Value: device,
	}
	// Use a device specific topic as we are the only core handling requests for this device
	replyToTopic := kafka.CreateSubTopic(ap.kafkaICProxy.DefaultTopic.Name, device.Id)
	success, result := ap.kafkaICProxy.InvokeRPC(ctx, rpc, &toTopic, &replyToTopic, true, args...)
	log.Debugw("RebootDevice-response", log.Fields{"deviceid": device.Id, "success": success})
	return unPackResponse(rpc, device.Id, success, result)
}

func (ap *AdapterProxy) DeleteDevice(ctx context.Context, device *voltha.Device) error {
	log.Debugw("DeleteDevice", log.Fields{"deviceId": device.Id})
	rpc := "delete_device"
	toTopic := kafka.CreateSubTopic(device.Type, device.Id)
	args := make([]*kafka.KVArg, 1)
	args[0] = &kafka.KVArg{
		Key:   "device",
		Value: device,
	}
	// Use a device specific topic as we are the only core handling requests for this device
	replyToTopic := kafka.CreateSubTopic(ap.kafkaICProxy.DefaultTopic.Name, device.Id)
	success, result := ap.kafkaICProxy.InvokeRPC(ctx, rpc, &toTopic, &replyToTopic, true, args...)
	log.Debugw("DeleteDevice-response", log.Fields{"deviceid": device.Id, "success": success})

	// We no longer need to have a target against that topic as we won't receive any unsolicited messages on that
	// topic
	if err := ap.kafkaICProxy.UnSubscribeFromRequestHandler(replyToTopic); err != nil {
		log.Errorw("Unable-to-subscribe-from-target", log.Fields{"topic": replyToTopic, "error": err})
		return err
	}
	// Now delete the topic altogether
	ap.kafkaICProxy.DeleteTopic(replyToTopic)

	return unPackResponse(rpc, device.Id, success, result)
}

func (ap *AdapterProxy) GetOfpDeviceInfo(ctx context.Context, device *voltha.Device) (*ic.SwitchCapability, error) {
	log.Debugw("GetOfpDeviceInfo", log.Fields{"deviceId": device.Id})
	toTopic := kafka.CreateSubTopic(device.Type, device.Id)
	args := make([]*kafka.KVArg, 1)
	args[0] = &kafka.KVArg{
		Key:   "device",
		Value: device,
	}
	// Use a device specific topic as we are the only core handling requests for this device
	replyToTopic := kafka.CreateSubTopic(ap.kafkaICProxy.DefaultTopic.Name, device.Id)
	success, result := ap.kafkaICProxy.InvokeRPC(ctx, "get_ofp_device_info", &toTopic, &replyToTopic, true, args...)
	log.Debugw("GetOfpDeviceInfo-response", log.Fields{"deviceId": device.Id, "success": success, "result": result})
	if success {
		unpackResult := &ic.SwitchCapability{}
		if err := ptypes.UnmarshalAny(result, unpackResult); err != nil {
			log.Warnw("cannot-unmarshal-response", log.Fields{"error": err})
			return nil, status.Errorf(codes.InvalidArgument, "%s", err.Error())
		}
		return unpackResult, nil
	} else {
		unpackResult := &ic.Error{}
		var err error
		if err = ptypes.UnmarshalAny(result, unpackResult); err != nil {
			log.Warnw("cannot-unmarshal-response", log.Fields{"error": err})
		}
		log.Debugw("GetOfpDeviceInfo-return", log.Fields{"deviceid": device.Id, "success": success, "error": err})
		// TODO:  Need to get the real error code
		return nil, status.Errorf(codes.Internal, "%s", unpackResult.Reason)
	}
}

func (ap *AdapterProxy) GetOfpPortInfo(ctx context.Context, device *voltha.Device, portNo uint32) (*ic.PortCapability, error) {
	log.Debugw("GetOfpPortInfo", log.Fields{"deviceId": device.Id})
	toTopic := kafka.CreateSubTopic(device.Type, device.Id)
	args := make([]*kafka.KVArg, 2)
	args[0] = &kafka.KVArg{
		Key:   "device",
		Value: device,
	}
	pNo := &ic.IntType{Val: int64(portNo)}
	args[1] = &kafka.KVArg{
		Key:   "port_no",
		Value: pNo,
	}
	// Use a device specific topic as we are the only core handling requests for this device
	replyToTopic := kafka.CreateSubTopic(ap.kafkaICProxy.DefaultTopic.Name, device.Id)
	success, result := ap.kafkaICProxy.InvokeRPC(ctx, "get_ofp_port_info", &toTopic, &replyToTopic, true, args...)
	log.Debugw("GetOfpPortInfo-response", log.Fields{"deviceid": device.Id, "success": success})
	if success {
		unpackResult := &ic.PortCapability{}
		if err := ptypes.UnmarshalAny(result, unpackResult); err != nil {
			log.Warnw("cannot-unmarshal-response", log.Fields{"error": err})
			return nil, status.Errorf(codes.InvalidArgument, "%s", err.Error())
		}
		return unpackResult, nil
	} else {
		unpackResult := &ic.Error{}
		var err error
		if err = ptypes.UnmarshalAny(result, unpackResult); err != nil {
			log.Warnw("cannot-unmarshal-response", log.Fields{"error": err})
		}
		log.Debugw("GetOfpPortInfo-return", log.Fields{"deviceid": device.Id, "success": success, "error": err})
		// TODO:  Need to get the real error code
		return nil, status.Errorf(codes.Internal, "%s", unpackResult.Reason)
	}
}

//TODO: Implement the functions below

func (ap *AdapterProxy) AdapterDescriptor() (*voltha.Adapter, error) {
	log.Debug("AdapterDescriptor")
	return nil, nil
}

func (ap *AdapterProxy) DeviceTypes() (*voltha.DeviceType, error) {
	log.Debug("DeviceTypes")
	return nil, nil
}

func (ap *AdapterProxy) Health() (*voltha.HealthStatus, error) {
	log.Debug("Health")
	return nil, nil
}

func (ap *AdapterProxy) ReconcileDevice(device *voltha.Device) error {
	log.Debug("ReconcileDevice")
	return nil
}

func (ap *AdapterProxy) AbandonDevice(device voltha.Device) error {
	log.Debug("AbandonDevice")
	return nil
}

func (ap *AdapterProxy) GetDeviceDetails(device voltha.Device) (*voltha.Device, error) {
	log.Debug("GetDeviceDetails")
	return nil, nil
}

func (ap *AdapterProxy) DownloadImage(device voltha.Device, download voltha.ImageDownload) error {
	log.Debug("DownloadImage")
	return nil
}

func (ap *AdapterProxy) GetImageDownloadStatus(device voltha.Device, download voltha.ImageDownload) error {
	log.Debug("GetImageDownloadStatus")
	return nil
}

func (ap *AdapterProxy) CancelImageDownload(device voltha.Device, download voltha.ImageDownload) error {
	log.Debug("CancelImageDownload")
	return nil
}

func (ap *AdapterProxy) ActivateImageUpdate(device voltha.Device, download voltha.ImageDownload) error {
	log.Debug("ActivateImageUpdate")
	return nil
}

func (ap *AdapterProxy) RevertImageUpdate(device voltha.Device, download voltha.ImageDownload) error {
	log.Debug("RevertImageUpdate")
	return nil
}

func (ap *AdapterProxy) SelfTestDevice(device voltha.Device) (*voltha.SelfTestResponse, error) {
	log.Debug("SelfTestDevice")
	return nil, nil
}

func (ap *AdapterProxy) packetOut(deviceType string, deviceId string, outPort uint32, packet *openflow_13.OfpPacketOut) error {
	log.Debugw("packetOut", log.Fields{"deviceId": deviceId})
	toTopic := kafka.CreateSubTopic(deviceType, deviceId)
	rpc := "receive_packet_out"
	dId := &ic.StrType{Val: deviceId}
	args := make([]*kafka.KVArg, 3)
	args[0] = &kafka.KVArg{
		Key:   "deviceId",
		Value: dId,
	}
	op := &ic.IntType{Val: int64(outPort)}
	args[1] = &kafka.KVArg{
		Key:   "outPort",
		Value: op,
	}
	args[2] = &kafka.KVArg{
		Key:   "packet",
		Value: packet,
	}

	// TODO:  Do we need to wait for an ACK on a packet Out?
	// Use a device specific topic as we are the only core handling requests for this device
	replyToTopic := kafka.CreateSubTopic(ap.kafkaICProxy.DefaultTopic.Name, deviceId)
	success, result := ap.kafkaICProxy.InvokeRPC(nil, rpc, &toTopic, &replyToTopic, false, args...)
	log.Debugw("packetOut", log.Fields{"deviceid": deviceId, "success": success})
	return unPackResponse(rpc, deviceId, success, result)
}

func (ap *AdapterProxy) UpdateFlowsBulk(device *voltha.Device, flows *voltha.Flows, groups *voltha.FlowGroups) error {
	log.Debugw("UpdateFlowsBulk", log.Fields{"deviceId": device.Id})
	toTopic := kafka.CreateSubTopic(device.Type, device.Id)
	rpc := "update_flows_bulk"
	args := make([]*kafka.KVArg, 3)
	args[0] = &kafka.KVArg{
		Key:   "device",
		Value: device,
	}
	args[1] = &kafka.KVArg{
		Key:   "flows",
		Value: flows,
	}
	args[2] = &kafka.KVArg{
		Key:   "groups",
		Value: groups,
	}

	// Use a device specific topic as we are the only core handling requests for this device
	replyToTopic := kafka.CreateSubTopic(ap.kafkaICProxy.DefaultTopic.Name, device.Id)
	success, result := ap.kafkaICProxy.InvokeRPC(nil, rpc, &toTopic, &replyToTopic, true, args...)
	log.Debugw("UpdateFlowsBulk-response", log.Fields{"deviceid": device.Id, "success": success})
	return unPackResponse(rpc, device.Id, success, result)
}

func (ap *AdapterProxy) UpdateFlowsIncremental(device *voltha.Device, flowChanges *openflow_13.FlowChanges, groupChanges *openflow_13.FlowGroupChanges) error {
	log.Debugw("UpdateFlowsIncremental", log.Fields{"deviceId": device.Id})
	toTopic := kafka.CreateSubTopic(device.Type, device.Id)
	rpc := "update_flows_bulk"
	args := make([]*kafka.KVArg, 3)
	args[0] = &kafka.KVArg{
		Key:   "device",
		Value: device,
	}
	args[1] = &kafka.KVArg{
		Key:   "flow_changes",
		Value: flowChanges,
	}
	args[2] = &kafka.KVArg{
		Key:   "group_changes",
		Value: groupChanges,
	}

	// Use a device specific topic as we are the only core handling requests for this device
	replyToTopic := kafka.CreateSubTopic(ap.kafkaICProxy.DefaultTopic.Name, device.Id)
	success, result := ap.kafkaICProxy.InvokeRPC(nil, rpc, &toTopic, &replyToTopic, true, args...)
	log.Debugw("UpdateFlowsIncremental-response", log.Fields{"deviceid": device.Id, "success": success})
	return unPackResponse(rpc, device.Id, success, result)
}

func (ap *AdapterProxy) UpdatePmConfig(device voltha.Device, pmConfigs voltha.PmConfigs) error {
	log.Debug("UpdatePmConfig")
	return nil
}

func (ap *AdapterProxy) ReceivePacketOut(deviceId voltha.ID, egressPortNo int, msg interface{}) error {
	log.Debug("ReceivePacketOut")
	return nil
}

func (ap *AdapterProxy) SuppressAlarm(filter voltha.AlarmFilter) error {
	log.Debug("SuppressAlarm")
	return nil
}

func (ap *AdapterProxy) UnSuppressAlarm(filter voltha.AlarmFilter) error {
	log.Debug("UnSuppressAlarm")
	return nil
}
