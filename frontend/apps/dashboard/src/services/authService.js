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
    localStorage.setItem(authTokenKey, token);
    return true;
  }
  return false;
}

export function passwordLogout() {
  localStorage.removeItem(authTokenKey);
}

export function checkAuthStatus() {
  const i = localStorage.getItem(authTokenKey);
  return Promise.resolve(!!i);
}
