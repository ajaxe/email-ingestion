export const apiEvents = new EventTarget();

export const API_EVENT_TYPES = {
  FORBIDDEN_TOKEN_INVALID: "api:forbidden_token_invalid",
  FORBIDDEN_USER: "api:forbidden_user",
};

export function emitForbiddenTokenInvalid() {
  apiEvents.dispatchEvent(new CustomEvent(API_EVENT_TYPES.FORBIDDEN_TOKEN_INVALID));
}

export function emitForbiddenUser() {
  apiEvents.dispatchEvent(new CustomEvent(API_EVENT_TYPES.FORBIDDEN_USER));
}
