export async function getUser() {
  const resp = await fetch("/app/userinfo");

  if (resp.status >= 200 && resp.status <= 299) {
    const u = await resp.json();
    return toUser(u);
  }

  const errorData = await resp.json().catch(() => ({}));
  const error = new Error(
    errorData.message || (resp.status === 401 ? "Unauthorized" : "Forbidden"),
  );
  error.code = errorData.code || "UNKNOWN";
  error.status = resp.status;
  throw error;
}

export async function getAuthSession() {
  const resp = await fetch("/app/v1/auth/session", {
    method: "GET",
  });

  if (resp.status >= 200 && resp.status <= 299) {
    return await resp.json();
  }

  const errorData = await resp.json().catch(() => ({}));
  const error = new Error(
    errorData.message || (resp.status === 401 ? "Unauthorized" : "Forbidden"),
  );
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
  const user = {
    name: "",
    email: "",
    profileImage: "",
    csrfToken: oidcUser.csrfToken,
    csrfFormFieldName: oidcUser.csrfFormFieldName,
  };

  const claimList = new Set(["name", "email", "picture"]);
  for (const claim of oidcUser.claims.filter((claim) =>
    claimList.has(claim.type),
  )) {
    if (claim.type === "name") {
      user.name = claim.value;
    }
    if (claim.type === "email") {
      user.email = claim.value;
    }
    if (claim.type === "picture") {
      user.profileImage = claim.value;
    }
  }

  return user;
}
