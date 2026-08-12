import { UserManager } from 'oidc-client-ts'
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
  scope: "openid profile email",

  automaticSilentRenew: true,
  loadUserInfo: true,
};
export const authSettings = {
  authCallbackUrl: loginCallbackUrl,
  logoutCallbackUrl,
};
console.log("authSettings", authSettings);
// Export the single instance
export const userManager = new UserManager(settings);

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
  } else if(window.APP_CONFIG.AUTH_PROVIDER === "oidc") {
    const u = await userManager.getUser()
    console.log("oidc user",u)
    return u
  }
  if (!err) {
    err = new Error("un-supported auth provider");
  }
  throw err;
}

export async function signoutRedirectCallback()  {
  return await userManager.signoutRedirectCallback();
}

export function signinRedirect() {
  return userManager.signinRedirect();
}

export async function signinRedirectCallback() {
  const u = await userManager.signinRedirectCallback();
  return {
    name: u.profile?.name,
    email: u.profile?.email,
    profileImage: u.profile?.picture,
    access_token: u.access_token,
  };
}
