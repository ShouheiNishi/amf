package consumer

import (
	"context"
	"fmt"
	"strings"

	"github.com/antihax/optional"

	amf_context "github.com/free5gc/amf/internal/context"
	"github.com/free5gc/amf/internal/logger"
	"github.com/free5gc/openapi"
	"github.com/free5gc/openapi/Nudm_SubscriberDataManagement"
	"github.com/free5gc/openapi/models"
)

func PutUpuAck(ue *amf_context.AmfUe, upuMacIue string) error {
	configuration := Nudm_SubscriberDataManagement.NewConfiguration()
	configuration.SetBasePath(ue.NudmSDMUri)
	client := Nudm_SubscriberDataManagement.NewAPIClient(configuration)

	ackInfo := models.AcknowledgeInfo{
		UpuMacIue: upuMacIue,
	}
	upuOpt := Nudm_SubscriberDataManagement.PutUpuAckParamOpts{
		AcknowledgeInfo: optional.NewInterface(ackInfo),
	}
	rsp, err := client.ProvidingAcknowledgementOfUEParametersUpdateApi.PutUpuAck(
		context.Background(), ue.Supi, &upuOpt)
	defer func() {
		if rsp != nil {
			if bodyCloseErr := rsp.Body.Close(); bodyCloseErr != nil {
				logger.ConsumerLog.Errorf("PutUpuAck' response body cannot close: %v", bodyCloseErr)
			}
		}
	}()
	return err
}

func SDMGetAmData(ue *amf_context.AmfUe) (problemDetails *models.ProblemDetails, err error) {
	configuration := Nudm_SubscriberDataManagement.NewConfiguration()
	configuration.SetBasePath(ue.NudmSDMUri)
	client := Nudm_SubscriberDataManagement.NewAPIClient(configuration)

	getAmDataParamOpt := Nudm_SubscriberDataManagement.GetAmDataParamOpts{
		PlmnId: optional.NewInterface(openapi.MarshToJsonString(ue.PlmnId)),
	}

	data, httpResp, localErr := client.AccessAndMobilitySubscriptionDataRetrievalApi.GetAmData(
		context.Background(), ue.Supi, &getAmDataParamOpt)
	if localErr == nil {
		ue.AccessAndMobilitySubscriptionData = &data
		ue.Gpsi = data.Gpsis[0] // TODO: select GPSI
	} else if httpResp != nil {
		defer func() {
			if bodyCloseErr := httpResp.Body.Close(); bodyCloseErr != nil {
				logger.ConsumerLog.Errorf("GetAmData' response body cannot close: %v", bodyCloseErr)
			}
		}()
		if httpResp.Status != localErr.Error() {
			err = localErr
			return
		}
		problem := localErr.(openapi.GenericOpenAPIError).Model().(models.ProblemDetails)
		problemDetails = &problem
	} else {
		err = openapi.ReportError("server no response")
	}
	return
}

func SDMGetSmfSelectData(ue *amf_context.AmfUe) (problemDetails *models.ProblemDetails, err error) {
	configuration := Nudm_SubscriberDataManagement.NewConfiguration()
	configuration.SetBasePath(ue.NudmSDMUri)
	client := Nudm_SubscriberDataManagement.NewAPIClient(configuration)

	paramOpt := Nudm_SubscriberDataManagement.GetSmfSelectDataParamOpts{
		PlmnId: optional.NewInterface(openapi.MarshToJsonString(ue.PlmnId)),
	}
	data, httpResp, localErr := client.SMFSelectionSubscriptionDataRetrievalApi.GetSmfSelectData(
		context.Background(), ue.Supi, &paramOpt)
	if localErr == nil {
		// SubscribedSnssaiInfos in SmfSelectionSubscriptionData requires case-insensitive search,
		// So keys are converted to lower case, here.
		if l := len(data.SubscribedSnssaiInfos); l != 0 {
			newSubscribedSnssaiInfos := make(map[string]models.SnssaiInfo, l)
			for key, info := range data.SubscribedSnssaiInfos {
				newKey := strings.ToLower(key)
				if _, exist := newSubscribedSnssaiInfos[newKey]; exist {
					return nil, fmt.Errorf("duplicate key %s", newKey)
				}
				newSubscribedSnssaiInfos[newKey] = info
			}
			data.SubscribedSnssaiInfos = newSubscribedSnssaiInfos
		}
		ue.SmfSelectionData = &data
	} else if httpResp != nil {
		defer func() {
			if bodyCloseErr := httpResp.Body.Close(); bodyCloseErr != nil {
				logger.ConsumerLog.Errorf("GetSmfSelectData' response body cannot close: %v", bodyCloseErr)
			}
		}()
		if httpResp.Status != localErr.Error() {
			err = localErr
			return
		}
		problem := localErr.(openapi.GenericOpenAPIError).Model().(models.ProblemDetails)
		problemDetails = &problem
	} else {
		err = openapi.ReportError("server no response")
	}

	return problemDetails, err
}

func SDMGetUeContextInSmfData(ue *amf_context.AmfUe) (problemDetails *models.ProblemDetails, err error) {
	configuration := Nudm_SubscriberDataManagement.NewConfiguration()
	configuration.SetBasePath(ue.NudmSDMUri)
	client := Nudm_SubscriberDataManagement.NewAPIClient(configuration)

	data, httpResp, localErr := client.UEContextInSMFDataRetrievalApi.GetUeContextInSmfData(
		context.Background(), ue.Supi, nil)
	if localErr == nil {
		ue.UeContextInSmfData = &data
	} else if httpResp != nil {
		defer func() {
			if bodyCloseErr := httpResp.Body.Close(); bodyCloseErr != nil {
				logger.ConsumerLog.Errorf("GetUeContextInSmfData' response body cannot close: %v", bodyCloseErr)
			}
		}()
		if httpResp.Status != localErr.Error() {
			err = localErr
			return
		}
		problem := localErr.(openapi.GenericOpenAPIError).Model().(models.ProblemDetails)
		problemDetails = &problem
	} else {
		err = openapi.ReportError("server no response")
	}

	return
}

func SDMSubscribe(ue *amf_context.AmfUe) (problemDetails *models.ProblemDetails, err error) {
	configuration := Nudm_SubscriberDataManagement.NewConfiguration()
	configuration.SetBasePath(ue.NudmSDMUri)
	client := Nudm_SubscriberDataManagement.NewAPIClient(configuration)

	amfSelf := amf_context.AMF_Self()
	sdmSubscription := models.SdmSubscription{
		NfInstanceId: amfSelf.NfId,
		PlmnId:       &ue.PlmnId,
	}

	resSubscription, httpResp, localErr := client.SubscriptionCreationApi.Subscribe(
		context.Background(), ue.Supi, sdmSubscription)
	if localErr == nil {
		ue.SdmSubscriptionId = resSubscription.SubscriptionId
		return
	} else if httpResp != nil {
		defer func() {
			if bodyCloseErr := httpResp.Body.Close(); bodyCloseErr != nil {
				logger.ConsumerLog.Errorf("Subscribe' response body cannot close: %v", bodyCloseErr)
			}
		}()
		if httpResp.Status != localErr.Error() {
			err = localErr
			return
		}
		problem := localErr.(openapi.GenericOpenAPIError).Model().(models.ProblemDetails)
		problemDetails = &problem
	} else {
		err = openapi.ReportError("server no response")
	}
	return problemDetails, err
}

func SDMGetSliceSelectionSubscriptionData(ue *amf_context.AmfUe) (problemDetails *models.ProblemDetails, err error) {
	configuration := Nudm_SubscriberDataManagement.NewConfiguration()
	configuration.SetBasePath(ue.NudmSDMUri)
	client := Nudm_SubscriberDataManagement.NewAPIClient(configuration)

	paramOpt := Nudm_SubscriberDataManagement.GetNssaiParamOpts{
		PlmnId: optional.NewInterface(openapi.MarshToJsonString(ue.PlmnId)),
	}
	nssai, httpResp, localErr := client.SliceSelectionSubscriptionDataRetrievalApi.GetNssai(
		context.Background(), ue.Supi, &paramOpt)
	if localErr == nil {
		for _, defaultSnssai := range nssai.DefaultSingleNssais {
			subscribedSnssai := models.SubscribedSnssai{
				SubscribedSnssai: &models.Snssai{
					Sst: defaultSnssai.Sst,
					Sd:  defaultSnssai.Sd,
				},
				DefaultIndication: true,
			}
			ue.SubscribedNssai = append(ue.SubscribedNssai, subscribedSnssai)
		}
		for _, snssai := range nssai.SingleNssais {
			subscribedSnssai := models.SubscribedSnssai{
				SubscribedSnssai: &models.Snssai{
					Sst: snssai.Sst,
					Sd:  snssai.Sd,
				},
				DefaultIndication: false,
			}
			ue.SubscribedNssai = append(ue.SubscribedNssai, subscribedSnssai)
		}
	} else if httpResp != nil {
		defer func() {
			if bodyCloseErr := httpResp.Body.Close(); bodyCloseErr != nil {
				logger.ConsumerLog.Errorf("GetNssai' response body cannot close: %v", bodyCloseErr)
			}
		}()
		if httpResp.Status != localErr.Error() {
			err = localErr
			return
		}
		problem := localErr.(openapi.GenericOpenAPIError).Model().(models.ProblemDetails)
		problemDetails = &problem
	} else {
		err = openapi.ReportError("server no response")
	}
	return problemDetails, err
}

func SDMUnsubscribe(ue *amf_context.AmfUe) (problemDetails *models.ProblemDetails, err error) {
	configuration := Nudm_SubscriberDataManagement.NewConfiguration()
	configuration.SetBasePath(ue.NudmSDMUri)
	client := Nudm_SubscriberDataManagement.NewAPIClient(configuration)

	httpResp, localErr := client.SubscriptionDeletionApi.Unsubscribe(context.Background(), ue.Supi, ue.SdmSubscriptionId)
	if localErr == nil {
		return
	} else if httpResp != nil {
		defer func() {
			if bodyCloseErr := httpResp.Body.Close(); bodyCloseErr != nil {
				logger.ConsumerLog.Errorf("Unsubscribe' response body cannot close: %v", bodyCloseErr)
			}
		}()
		if httpResp.Status != localErr.Error() {
			err = localErr
			return
		}
		problem := localErr.(openapi.GenericOpenAPIError).Model().(models.ProblemDetails)
		problemDetails = &problem
	} else {
		err = openapi.ReportError("server no response")
	}
	return
}
