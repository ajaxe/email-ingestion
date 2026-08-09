/* eslint-disable antfu/top-level-function */
import axios from "axios";
// import { UserManager } from "oidc-client-ts";
import { fetchToken } from "./authService";

const apiClient = axios.create({
  baseURL: "/app/v1",
});

/* async function oidcTokenProvider(silentSignin = false) {
  if (silentSignin) {
    await UserManager.signinSilent();
  }
  const user = await UserManager.getUser();
  return user.access_token;
} */

async function passwordTokenProvider() {
  const { access_token } = await fetchToken();
  return access_token;
}

const tokenProvider = passwordTokenProvider;

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

export const createApiKey = (appId, name = "Default API Key") =>
  apiClient.post(`/applications/${appId}/api-keys`, { name });

export const getApplicationStats = (appId) => apiClient.get(`/applications/${appId}/stats`);
