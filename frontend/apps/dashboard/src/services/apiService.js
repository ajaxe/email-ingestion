import axios from "axios";
import { tokenProvider } from "./authService";
import { emitForbiddenTokenInvalid, emitForbiddenUser } from "./apiEvents";

const apiClient = axios.create({
  baseURL: "/app/v1",
});

apiClient.interceptors.request.use(async (config) => {
  const token = await tokenProvider();
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

apiClient.interceptors.response.use(
  (response) => response,
  async (error) => {
    if (error.response?.status === 401) {
      try {
        const token = await tokenProvider(true);
        error.config.headers.Authorization = `Bearer ${token}`;
        return apiClient(error.config);
      } catch (refreshErr) {
        emitForbiddenTokenInvalid();
        throw refreshErr;
      }
    }

    if (error.response?.status === 403) {
      const errorMessage = error.response?.data?.message;
      if (errorMessage === "forbidden_token_invalid") {
        emitForbiddenTokenInvalid();
      } else if (errorMessage === "forbidden_user") {
        emitForbiddenUser();
      }
    }

    throw error;
  },
);

export function getApplicationById (appId) {
  return apiClient.get(`/applications/${appId}`)
}

export const getApplications = () => apiClient.get(`/applications`);

export function createApplication (payload) {
  return apiClient.post(`/applications`, typeof payload === 'string' ? { name: payload } : payload)
}

export function getAddresses (appId, queryParams) {
  return apiClient.get(`/applications/${appId}/addresses`, { params: queryParams })
}

export function createAddress (appId, description) {
  return apiClient.post(`/applications/${appId}/addresses`, { description })
}

export function toggleAddressStatus (appId, addressId, status) {
  return apiClient.patch(`/applications/${appId}/addresses/${addressId}`, { status })
}

export function getEmailList (appId, queryParams) {
  return apiClient.get(`/applications/${appId}/emails`, { params: queryParams })
}

export function getEmailById (appId, emailId) {
  return apiClient.get(`/applications/${appId}/emails/${emailId}`)
}

export function deleteEmail (appId, emailId) {
  return apiClient.delete(`/applications/${appId}/emails/${emailId}`)
}

export function bulkDeleteEmails (appId, emailIds) {
  return apiClient.post(`/applications/${appId}/emails/bulk-delete`, { emailIds })
}

export function getEmailWebhookHistory (appId, emailId) {
  return apiClient.get(`/applications/${appId}/emails/${emailId}/webhooks`)
}

export function getAttachmentUrl (appId, emailId, attachmentId) {
  return apiClient.get(
    `/applications/${appId}/emails/${emailId}/attachments/${attachmentId}`,
  )
}

export function registerWebhook (appId, config) {
  return apiClient.post(`/applications/${appId}/webhook`, config)
}

export function updateWebhook (appId, config) {
  return apiClient.put(`/applications/${appId}/webhook`, config)
}

export const configureWebhook = updateWebhook;

export function getWebhookJobs (appId, queryParams) {
  return apiClient.get(`/applications/${appId}/webhook/jobs`, { params: queryParams })
}

export function redeliverWebhook (appId, jobId) {
  return apiClient.post(`/applications/${appId}/webhook/jobs/${jobId}/redeliver`)
}

export function getApiKeys (appId) {
  return apiClient.get(`/applications/${appId}/api-keys`)
}

export function createApiKey (appId, payload) {
  return apiClient.post(`/applications/${appId}/api-keys`, typeof payload === 'string' ? { name: payload } : payload)
}

export function revokeApiKey (appId, keyId) {
  return apiClient.delete(`/applications/${appId}/api-keys/${keyId}`)
}

export const getApplicationStats = (appId) => apiClient.get(`/applications/${appId}/stats`);
