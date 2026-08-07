/* eslint-disable antfu/top-level-function */
import axios from "axios";
// import { UserManager } from "oidc-client-ts";
import { fetchToken} from "./authServiceauth";

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
    return error;
  },
);

export const getApplication = (appId) =>
  apiClient.get(`/applications/${appId}`);
export const createAddress = (appId, description) =>
  apiClient.post(`/applications/${appId}/addresses`, { description });
export const toggleAddressStatus = (appId, addressId, status) =>
  apiClient.patch(`/applications/${appId}/addresses/${addressId}`, { status });
export const getEmailList = (appId, queryParams) =>
  apiClient.get(`/applications/${appId}/emails`, { params: queryParams });
export const getAttachmentUrl = (appId, emailId, attachmentId) =>
  apiClient.get(
    `/applications/${appId}/emails/${emailId}/attachments/${attachmentId}`,
  );
export const configureWebhook = (appId, config) =>
  apiClient.put(`/applications/${appId}/webhook`, config);
export const getWebhookJobs = (appId, queryParams) =>
  apiClient.get(`/applications/${appId}/webhook/jobs`, { params: queryParams });
export const redeliverWebhook = (appId, jobId) =>
  apiClient.post(`/applications/${appId}/webhook/jobs/${jobId}/redeliver`);
