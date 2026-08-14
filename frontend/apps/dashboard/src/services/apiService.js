/* eslint-disable antfu/top-level-function */
import axios from "axios";
import { tokenProvider } from "./authService";

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
      const token = await tokenProvider(true);
      error.config.headers.Authorization = `Bearer ${token}`;
      return apiClient(error.config);
    }
    throw error;
  },
);

export const getApplicationById = (appId) =>
  apiClient.get(`/applications/${appId}`);

export const getApplications = () => apiClient.get(`/applications`);

export const createApplication = (payload) =>
  apiClient.post(`/applications`, typeof payload === 'string' ? { name: payload } : payload);

export const getAddresses = (appId, queryParams) =>
  apiClient.get(`/applications/${appId}/addresses`, { params: queryParams });

export const createAddress = (appId, description) =>
  apiClient.post(`/applications/${appId}/addresses`, { description });

export const toggleAddressStatus = (appId, addressId, status) =>
  apiClient.patch(`/applications/${appId}/addresses/${addressId}`, { status });

export const getEmailList = (appId, queryParams) =>
  apiClient.get(`/applications/${appId}/emails`, { params: queryParams });

export const getEmailById = (appId, emailId) =>
  apiClient.get(`/applications/${appId}/emails/${emailId}`);

export const getAttachmentUrl = (appId, emailId, attachmentId) =>
  apiClient.get(
    `/applications/${appId}/emails/${emailId}/attachments/${attachmentId}`,
  );

export const registerWebhook = (appId, config) =>
  apiClient.post(`/applications/${appId}/webhook`, config);

export const updateWebhook = (appId, config) =>
  apiClient.put(`/applications/${appId}/webhook`, config);

export const configureWebhook = updateWebhook;

export const getWebhookJobs = (appId, queryParams) =>
  apiClient.get(`/applications/${appId}/webhook/jobs`, { params: queryParams });

export const redeliverWebhook = (appId, jobId) =>
  apiClient.post(`/applications/${appId}/webhook/jobs/${jobId}/redeliver`);

export const getApiKeys = (appId) =>
  apiClient.get(`/applications/${appId}/api-keys`);

export const createApiKey = (appId, payload) =>
  apiClient.post(`/applications/${appId}/api-keys`, typeof payload === 'string' ? { name: payload } : payload);

export const revokeApiKey = (appId, keyId) =>
  apiClient.delete(`/applications/${appId}/api-keys/${keyId}`);

export const getApplicationStats = (appId) => apiClient.get(`/applications/${appId}/stats`);
