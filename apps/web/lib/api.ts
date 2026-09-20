
export const API_URL =
  process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export class ApiError extends Error {
  code: string;
  status: number;
  constructor(message: string, code: string, status: number) {
    super(message);
    this.code = code;
    this.status = status;
  }
}

let refreshing: Promise<boolean> | null = null;

/**
 * rpc calls /halomail.<service>.v1.<Service>/<Method>.
 * @param procedure e.g. "halomail.identity.v1.AuthService/Login"
 */
export async function rpc<T = unknown>(
  procedure: string,
  body: unknown = {},
  token?: string | null,
  accessKey?: string,
): Promise<T> {
  const request = () => fetch(`/api/rpc/${procedure}`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      ...(accessKey ? { "X-HaloMail-Key": accessKey } : {}),
    },
    body: JSON.stringify(body ?? {}),
  });
  void token;
  let res = await request();
  if (res.status === 401 && !procedure.includes("AuthService/Login") && !procedure.includes("AuthService/Register") && !procedure.includes("AuthService/RefreshSession") && !accessKey) {
    refreshing ??= fetch("/api/rpc/halomail.identity.v1.AuthService/RefreshSession", {
      method: "POST", headers: { "Content-Type": "application/json" }, body: "{}",
    }).then(response => response.ok).finally(() => { refreshing = null; });
    if (await refreshing) res = await request();
  }

  if (!res.ok) {
    let message = res.statusText;
    let code = "unknown";
    try {
      const j = await res.json();
      message = j.message || message;
      code = j.code || code;
    } catch {
      /* non-JSON error */
    }
    throw new ApiError(message, code, res.status);
  }
  return (await res.json()) as T;
}
