package nas

//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
//
// Code generated by Alibaba Cloud SDK Code Generator.
// Changes may cause incorrect behavior and will be lost if the code is regenerated.

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
)

// DisableSmbAcl invokes the nas.DisableSmbAcl API synchronously
func (client *Client) DisableSmbAcl(request *DisableSmbAclRequest) (response *DisableSmbAclResponse, err error) {
	response = CreateDisableSmbAclResponse()
	err = client.DoAction(request, response)
	return
}

// DisableSmbAclWithChan invokes the nas.DisableSmbAcl API asynchronously
func (client *Client) DisableSmbAclWithChan(request *DisableSmbAclRequest) (<-chan *DisableSmbAclResponse, <-chan error) {
	responseChan := make(chan *DisableSmbAclResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.DisableSmbAcl(request)
		if err != nil {
			errChan <- err
		} else {
			responseChan <- response
		}
	})
	if err != nil {
		errChan <- err
		close(responseChan)
		close(errChan)
	}
	return responseChan, errChan
}

// DisableSmbAclWithCallback invokes the nas.DisableSmbAcl API asynchronously
func (client *Client) DisableSmbAclWithCallback(request *DisableSmbAclRequest, callback func(response *DisableSmbAclResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *DisableSmbAclResponse
		var err error
		defer close(result)
		response, err = client.DisableSmbAcl(request)
		callback(response, err)
		result <- 1
	})
	if err != nil {
		defer close(result)
		callback(nil, err)
		result <- 0
	}
	return result
}

// DisableSmbAclRequest is the request struct for api DisableSmbAcl
type DisableSmbAclRequest struct {
	*requests.RpcRequest
	FileSystemId string `position:"Query" name:"FileSystemId"`
}

// DisableSmbAclResponse is the response struct for api DisableSmbAcl
type DisableSmbAclResponse struct {
	*responses.BaseResponse
	RequestId string `json:"RequestId" xml:"RequestId"`
}

// CreateDisableSmbAclRequest creates a request to invoke DisableSmbAcl API
func CreateDisableSmbAclRequest() (request *DisableSmbAclRequest) {
	request = &DisableSmbAclRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("NAS", "2017-06-26", "DisableSmbAcl", "nas", "openAPI")
	request.Method = requests.POST
	return
}

// CreateDisableSmbAclResponse creates a response to parse from DisableSmbAcl response
func CreateDisableSmbAclResponse() (response *DisableSmbAclResponse) {
	response = &DisableSmbAclResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
