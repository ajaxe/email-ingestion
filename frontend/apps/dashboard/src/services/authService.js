import { UserManager } from "oidc-client-ts";
import { jwtDecode } from "jwt-decode";

const authTokenKey = "authToken";

const loginCallbackUrl = `${location.origin}/auth/callback`;
const logoutCallbackUrl = `${location.origin}/auth/logout`;

/**
 * @type {import('oidc-client-ts').UserManagerSettings}
 */
const settings = {
  authority: window.APP_CONFIG?.OIDC_AUTHORITY,
  client_id: window.APP_CONFIG?.OIDC_CLIENT_ID,

  // The route in your Vue app that will handle the callback
  redirect_uri: loginCallbackUrl,

  // The route to return to after logging out
  post_logout_redirect_uri: logoutCallbackUrl,

  // **This enables the Auth Code + PKCE flow**
  response_type: "code",
  scope: "openid profile email offline_access",

  automaticSilentRenew: true,
  loadUserInfo: true,
};
export const authSettings = {
  authCallbackUrl: loginCallbackUrl,
  logoutCallbackUrl,
};
// Export the single instance
export const userManager = new UserManager(settings);

async function oidcTokenProvider(silentSignin = false) {
  if (silentSignin) {
    await userManager.signinSilent();
  }
  const user = await userManager.getUser();
  return user?.access_token;
}

async function passwordTokenProvider() {
  const { access_token } = await getUser();
  return access_token;
}

export const tokenProvider = window.APP_CONFIG.AUTH_PROVIDER === "password" ? passwordTokenProvider : oidcTokenProvider;

export async function passwordLogin(username, password) {
  const resp = await fetch("/app/v1/auth/login", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ username, password }),
  });

  if (!resp.ok) {
    const err = await resp.json();
    throw new Error(err.message);
  }

  const { token } = await resp.json();

  if (token) {
    sessionStorage.setItem(authTokenKey, token);
    return true;
  }
  return false;
}

export function passwordLogout() {
  sessionStorage.removeItem(authTokenKey);
}

export async function getUser() {
  let err = null;
  if (window.APP_CONFIG.AUTH_PROVIDER === "password") {
    try {
      const token = sessionStorage.getItem(authTokenKey);
      const decodedPayload = jwtDecode(token);
      return {
        name: decodedPayload.username,
        email: decodedPayload.email,
        profileImage: "",
        access_token: token,
      };
    } catch (error) {
      err = error;
    }
  } else if (window.APP_CONFIG.AUTH_PROVIDER === "oidc") {
    const u = await userManager.getUser();
    console.log("oidc user", toUser(u));
    return toUser(u);
  }
  if (!err) {
    err = new Error("un-supported auth provider");
  }
  throw err;
}

export async function signoutRedirectCallback() {
  return await userManager.signoutRedirectCallback();
}

export function signoutRedirect() {
  return userManager.signoutRedirect();
}

export function signinRedirect() {
  return userManager.signinRedirect();
}

export async function signinRedirectCallback() {
  const u = await userManager.signinRedirectCallback();
  return toUser(u);
}

export async function getAuthSession() {
  const token = await tokenProvider();
  const resp = await fetch("/app/v1/auth/session", {
    method: "GET",
    headers: {
      "Authorization": "Bearer " + token,
      "Content-Type": "application/json",
    },
  });

  if (resp.status >= 200 && resp.status <= 299) {
    return await resp.json();
  }

  const errorData = await resp.json().catch(() => ({}));
  const error = new Error(errorData.message || (resp.status === 401 ? "Unauthorized" : "Forbidden"));
  error.code = errorData.code || "UNKNOWN";
  error.status = resp.status;
  throw error;
}

export const emptyUser = {
  name: "",
  email: "",
  profileImage: "",
};

function toUser(oidcUser) {
  if (!oidcUser) return emptyUser;

  return {
    name: oidcUser.profile?.name,
    email: oidcUser.profile?.email,
    profileImage: oidcUser.profile?.picture,
    access_token: oidcUser.access_token,
  };
}
