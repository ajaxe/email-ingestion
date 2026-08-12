import { jwtDecode } from "jwt-decode";

const authTokenKey = "authToken";
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

export async function checkAuthStatus() {
  const { access_token } = await fetchToken();
  return !!access_token;
}

export function fetchToken() {
  return Promise.resolve({
    access_token: sessionStorage.getItem(authTokenKey),
  });
}

export async function getUserProfile() {
  let err = null;
  if (window.APP_CONFIG.AUTH_PROVIDER === "password") {
    try {
      const { access_token: token } = await fetchToken();
      const decodedPayload = jwtDecode(token);
      return {
        name: decodedPayload.username,
        email: decodedPayload.email,
        profileImage: "",
      };
    } catch (error) {
      err = error;
    }
  }
  if (!err) {
    err = new Error("un-supported auth provider");
  }
  throw err;
}
