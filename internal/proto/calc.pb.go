// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v5.29.3
// source: internal/proto/calc.proto

package calc

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type GetTaskRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetTaskRequest) Reset() {
	*x = GetTaskRequest{}
	mi := &file_internal_proto_calc_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetTaskRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetTaskRequest) ProtoMessage() {}

func (x *GetTaskRequest) ProtoReflect() protoreflect.Message {
	mi := &file_internal_proto_calc_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetTaskRequest.ProtoReflect.Descriptor instead.
func (*GetTaskRequest) Descriptor() ([]byte, []int) {
	return file_internal_proto_calc_proto_rawDescGZIP(), []int{0}
}

// Ответ
type GetTaskResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Status        string                 `protobuf:"bytes,1,opt,name=status,proto3" json:"status,omitempty"`
	Task          *TaskData              `protobuf:"bytes,2,opt,name=task,proto3" json:"task,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetTaskResponse) Reset() {
	*x = GetTaskResponse{}
	mi := &file_internal_proto_calc_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetTaskResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetTaskResponse) ProtoMessage() {}

func (x *GetTaskResponse) ProtoReflect() protoreflect.Message {
	mi := &file_internal_proto_calc_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetTaskResponse.ProtoReflect.Descriptor instead.
func (*GetTaskResponse) Descriptor() ([]byte, []int) {
	return file_internal_proto_calc_proto_rawDescGZIP(), []int{1}
}

func (x *GetTaskResponse) GetStatus() string {
	if x != nil {
		return x.Status
	}
	return ""
}

func (x *GetTaskResponse) GetTask() *TaskData {
	if x != nil {
		return x.Task
	}
	return nil
}

// TaskData описывает задачу
type TaskData struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            int32                  `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Arg1          float64                `protobuf:"fixed64,2,opt,name=arg1,proto3" json:"arg1,omitempty"`
	Arg2          float64                `protobuf:"fixed64,3,opt,name=arg2,proto3" json:"arg2,omitempty"`
	Operation     string                 `protobuf:"bytes,4,opt,name=operation,proto3" json:"operation,omitempty"`
	OperationTime int32                  `protobuf:"varint,5,opt,name=operation_time,json=operationTime,proto3" json:"operation_time,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *TaskData) Reset() {
	*x = TaskData{}
	mi := &file_internal_proto_calc_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TaskData) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TaskData) ProtoMessage() {}

func (x *TaskData) ProtoReflect() protoreflect.Message {
	mi := &file_internal_proto_calc_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TaskData.ProtoReflect.Descriptor instead.
func (*TaskData) Descriptor() ([]byte, []int) {
	return file_internal_proto_calc_proto_rawDescGZIP(), []int{2}
}

func (x *TaskData) GetId() int32 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *TaskData) GetArg1() float64 {
	if x != nil {
		return x.Arg1
	}
	return 0
}

func (x *TaskData) GetArg2() float64 {
	if x != nil {
		return x.Arg2
	}
	return 0
}

func (x *TaskData) GetOperation() string {
	if x != nil {
		return x.Operation
	}
	return ""
}

func (x *TaskData) GetOperationTime() int32 {
	if x != nil {
		return x.OperationTime
	}
	return 0
}

type PostResultRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            int32                  `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Result        float64                `protobuf:"fixed64,2,opt,name=result,proto3" json:"result,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *PostResultRequest) Reset() {
	*x = PostResultRequest{}
	mi := &file_internal_proto_calc_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PostResultRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PostResultRequest) ProtoMessage() {}

func (x *PostResultRequest) ProtoReflect() protoreflect.Message {
	mi := &file_internal_proto_calc_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PostResultRequest.ProtoReflect.Descriptor instead.
func (*PostResultRequest) Descriptor() ([]byte, []int) {
	return file_internal_proto_calc_proto_rawDescGZIP(), []int{3}
}

func (x *PostResultRequest) GetId() int32 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *PostResultRequest) GetResult() float64 {
	if x != nil {
		return x.Result
	}
	return 0
}

type PostResultResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Status        string                 `protobuf:"bytes,1,opt,name=status,proto3" json:"status,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *PostResultResponse) Reset() {
	*x = PostResultResponse{}
	mi := &file_internal_proto_calc_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PostResultResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PostResultResponse) ProtoMessage() {}

func (x *PostResultResponse) ProtoReflect() protoreflect.Message {
	mi := &file_internal_proto_calc_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PostResultResponse.ProtoReflect.Descriptor instead.
func (*PostResultResponse) Descriptor() ([]byte, []int) {
	return file_internal_proto_calc_proto_rawDescGZIP(), []int{4}
}

func (x *PostResultResponse) GetStatus() string {
	if x != nil {
		return x.Status
	}
	return ""
}

var File_internal_proto_calc_proto protoreflect.FileDescriptor

const file_internal_proto_calc_proto_rawDesc = "" +
	"\n" +
	"\x19internal/proto/calc.proto\x12\x04calc\"\x10\n" +
	"\x0eGetTaskRequest\"M\n" +
	"\x0fGetTaskResponse\x12\x16\n" +
	"\x06status\x18\x01 \x01(\tR\x06status\x12\"\n" +
	"\x04task\x18\x02 \x01(\v2\x0e.calc.TaskDataR\x04task\"\x87\x01\n" +
	"\bTaskData\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\x05R\x02id\x12\x12\n" +
	"\x04arg1\x18\x02 \x01(\x01R\x04arg1\x12\x12\n" +
	"\x04arg2\x18\x03 \x01(\x01R\x04arg2\x12\x1c\n" +
	"\toperation\x18\x04 \x01(\tR\toperation\x12%\n" +
	"\x0eoperation_time\x18\x05 \x01(\x05R\roperationTime\";\n" +
	"\x11PostResultRequest\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\x05R\x02id\x12\x16\n" +
	"\x06result\x18\x02 \x01(\x01R\x06result\",\n" +
	"\x12PostResultResponse\x12\x16\n" +
	"\x06status\x18\x01 \x01(\tR\x06status2\x86\x01\n" +
	"\vCalcService\x126\n" +
	"\aGetTask\x12\x14.calc.GetTaskRequest\x1a\x15.calc.GetTaskResponse\x12?\n" +
	"\n" +
	"PostResult\x12\x17.calc.PostResultRequest\x1a\x18.calc.PostResultResponseB?Z=github.com/TuHeKocmoc/yalyceumfinal2/internal/proto/calc;calcb\x06proto3"

var (
	file_internal_proto_calc_proto_rawDescOnce sync.Once
	file_internal_proto_calc_proto_rawDescData []byte
)

func file_internal_proto_calc_proto_rawDescGZIP() []byte {
	file_internal_proto_calc_proto_rawDescOnce.Do(func() {
		file_internal_proto_calc_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_internal_proto_calc_proto_rawDesc), len(file_internal_proto_calc_proto_rawDesc)))
	})
	return file_internal_proto_calc_proto_rawDescData
}

var file_internal_proto_calc_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_internal_proto_calc_proto_goTypes = []any{
	(*GetTaskRequest)(nil),     // 0: calc.GetTaskRequest
	(*GetTaskResponse)(nil),    // 1: calc.GetTaskResponse
	(*TaskData)(nil),           // 2: calc.TaskData
	(*PostResultRequest)(nil),  // 3: calc.PostResultRequest
	(*PostResultResponse)(nil), // 4: calc.PostResultResponse
}
var file_internal_proto_calc_proto_depIdxs = []int32{
	2, // 0: calc.GetTaskResponse.task:type_name -> calc.TaskData
	0, // 1: calc.CalcService.GetTask:input_type -> calc.GetTaskRequest
	3, // 2: calc.CalcService.PostResult:input_type -> calc.PostResultRequest
	1, // 3: calc.CalcService.GetTask:output_type -> calc.GetTaskResponse
	4, // 4: calc.CalcService.PostResult:output_type -> calc.PostResultResponse
	3, // [3:5] is the sub-list for method output_type
	1, // [1:3] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_internal_proto_calc_proto_init() }
func file_internal_proto_calc_proto_init() {
	if File_internal_proto_calc_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_internal_proto_calc_proto_rawDesc), len(file_internal_proto_calc_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_internal_proto_calc_proto_goTypes,
		DependencyIndexes: file_internal_proto_calc_proto_depIdxs,
		MessageInfos:      file_internal_proto_calc_proto_msgTypes,
	}.Build()
	File_internal_proto_calc_proto = out.File
	file_internal_proto_calc_proto_goTypes = nil
	file_internal_proto_calc_proto_depIdxs = nil
}
