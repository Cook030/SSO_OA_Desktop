import { type UseMutationOptions, type UseQueryOptions, useMutation, useQuery } from "@tanstack/react-query";
import { GraphQLClient } from "graphql-request";
import { getSdk as getDmpSdk } from "../generated-ts/service-dmp/sdk";
import { getSdk as getSso2Sdk, type SdkFunctionWrapper } from "../generated-ts/service-sso2/sdk";

// ---------------------------------------------------------------------------
// 包装层数据结构 & 通用工具
// ---------------------------------------------------------------------------

type WrappedGraphQLPayload = {
  code?: number;
  msg?: string;
  data?: unknown;
};

type GraphQLExecutionResult = {
  data?: Record<string, unknown>;
  errors?: Array<{ message?: string }>;
};

const sso2Endpoint = import.meta.env.VITE_SSO2_GRAPHQL_ENDPOINT;
const dmpEndpoint = import.meta.env.VITE_DMP_GRAPHQL_ENDPOINT;

let refreshPromise: Promise<void> | null = null;
let hasRedirectedToException = false;

function redirectToExceptionPage(path: "/exception/403" | "/exception/500"): void {
  if (typeof window === "undefined" || hasRedirectedToException) {
    return;
  }

  const { pathname } = window.location;
  if (pathname.includes(path)) {
    return;
  }

  hasRedirectedToException = true;
  const basePath = pathname.includes("/dsp/") || pathname.endsWith("/dsp") ? "/dsp" : "";
  window.location.assign(`${basePath}${path}`);
}

function parseWrappedPayload(value: unknown): WrappedGraphQLPayload | null {
  if (!value) {
    return null;
  }

  try {
    return typeof value === "string" ? (JSON.parse(value) as WrappedGraphQLPayload) : (value as WrappedGraphQLPayload);
  } catch {
    return null;
  }
}

async function readExecutionResult(response: Response): Promise<GraphQLExecutionResult | null> {
  try {
    return (await response.clone().json()) as GraphQLExecutionResult;
  } catch {
    return null;
  }
}

function isRefreshOperation(init?: RequestInit): boolean {
  if (typeof init?.body !== "string") {
    return false;
  }

  return init.body.includes("mhSso2_authRefresh");
}

async function isAuthExpiredResponse(response: Response): Promise<boolean> {
  const result = await readExecutionResult(response);
  if (!result?.data) {
    return false;
  }

  return Object.values(result.data).some(value => parseWrappedPayload(value)?.code === 401);
}

async function handleBusinessErrorRedirect(response: Response): Promise<void> {
  if (response.status === 403) {
    redirectToExceptionPage("/exception/403");
    return;
  }

  if (response.status >= 500) {
    redirectToExceptionPage("/exception/500");
    return;
  }
}

async function refreshSession(): Promise<void> {
  const response = await fetch(sso2Endpoint, {
    method: "POST",
    credentials: "include",
    headers: {
      "Content-Type": "application/json"
    },
    body: JSON.stringify({
      query: `mutation RefreshSession($input: mhSso2_mutationInput_authRefresh_input_Input) {
        mhSso2_authRefresh(input: $input)
      }`,
      variables: {
        input: {}
      }
    })
  });

  const result = await readExecutionResult(response);
  const payload = parseWrappedPayload(result?.data?.mhSso2_authRefresh);
  if (payload?.code !== 200) {
    throw new Error(payload?.msg || "failed to refresh session");
  }
}

async function ensureSessionRefreshed(): Promise<void> {
  if (!refreshPromise) {
    refreshPromise = refreshSession().finally(() => {
      refreshPromise = null;
    });
  }

  await refreshPromise;
}

async function authAwareFetch(input: RequestInfo | URL, init?: RequestInit): Promise<Response> {
  const requestInit = {
    ...init,
    credentials: "include" as const
  };

  const response = await fetch(input, requestInit);
  if (isRefreshOperation(init)) {
    return response;
  }

  if (!(await isAuthExpiredResponse(response))) {
    await handleBusinessErrorRedirect(response);
    return response;
  }

  await ensureSessionRefreshed();
  const retryResponse = await fetch(input, requestInit);
  await handleBusinessErrorRedirect(retryResponse);
  return retryResponse;
}

// ---------------------------------------------------------------------------
// 业务异常 & 解包 wrapper
// ---------------------------------------------------------------------------

/** 业务层 GraphQL 异常：当包装层 code !== 200 时抛出 */
export class GraphqlBizError extends Error {
  readonly code: number;
  readonly msg: string;
  readonly operation: string;

  constructor(code: number, msg: string, operation: string) {
    super(`[${operation}] ${code}: ${msg}`);
    this.name = "GraphqlBizError";
    this.code = code;
    this.msg = msg;
    this.operation = operation;
  }
}

/**
 * 统一拆掉 `{code, msg, data}` 包装层：
 * - code === 200：把字段值替换为 data
 * - code !== 200：抛 GraphqlBizError，并按需触发跳转
 * - 没有包装结构的字段保持原值
 */
const unwrapWrapper: SdkFunctionWrapper = async (action, operationName) => {
  const raw = (await action()) as Record<string, unknown>;
  if (!raw || typeof raw !== "object") {
    return raw as never;
  }

  const unwrapped: Record<string, unknown> = {};
  for (const [field, value] of Object.entries(raw)) {
    const payload = parseWrappedPayload(value);
    if (payload && typeof payload === "object" && "code" in payload) {
      const { code, msg, data } = payload;
      if (code !== 200) {
        if (code === 403) redirectToExceptionPage("/exception/403");
        else if (code && code >= 500) redirectToExceptionPage("/exception/500");
        throw new GraphqlBizError(code ?? -1, msg ?? "未知错误", operationName);
      }
      unwrapped[field] = data;
    } else {
      unwrapped[field] = value;
    }
  }
  return unwrapped as never;
};

// ---------------------------------------------------------------------------
// 双服务 SDK
// ---------------------------------------------------------------------------

const sso2Client = new GraphQLClient(sso2Endpoint, { fetch: authAwareFetch });
const dmpClient = new GraphQLClient(dmpEndpoint, { fetch: authAwareFetch });

export const sdk = {
  sso2: getSso2Sdk(sso2Client, unwrapWrapper),
  dmp: getDmpSdk(dmpClient, unwrapWrapper)
};

export type AppSdk = typeof sdk;
export type SdkNamespace = keyof AppSdk;

// ---------------------------------------------------------------------------
// React Query Hooks（自动绑定 sdk 方法）
// ---------------------------------------------------------------------------

type SdkMethod = (variables?: any, headers?: any, signal?: any) => Promise<any>;
type MethodOf<N extends SdkNamespace> = {
  [K in keyof AppSdk[N]]: AppSdk[N][K] extends SdkMethod ? K : never;
}[keyof AppSdk[N]];

type VariablesOf<N extends SdkNamespace, K extends MethodOf<N>> = AppSdk[N][K] extends (
  v?: infer V,
  ...rest: any[]
) => any
  ? V
  : never;

type ResultOf<N extends SdkNamespace, K extends MethodOf<N>> = AppSdk[N][K] extends (...args: any[]) => Promise<infer R>
  ? R
  : never;

/**
 * 自动绑定 sdk 方法的查询 hook，统一约定 queryKey: [namespace, method, variables]
 */
export function useSdkQuery<N extends SdkNamespace, K extends MethodOf<N>>(
  ns: N,
  method: K,
  variables?: VariablesOf<N, K>,
  options?: Omit<UseQueryOptions<ResultOf<N, K>, Error>, "queryKey" | "queryFn">
) {
  return useQuery<ResultOf<N, K>, Error>({
    queryKey: [ns, method, variables],
    queryFn: ({ signal }) => (sdk[ns][method] as SdkMethod)(variables, undefined, signal),
    ...options
  });
}

/** 自动绑定 sdk 方法的 mutation hook */
export function useSdkMutation<N extends SdkNamespace, K extends MethodOf<N>>(
  ns: N,
  method: K,
  options?: UseMutationOptions<ResultOf<N, K>, Error, VariablesOf<N, K>>
) {
  return useMutation<ResultOf<N, K>, Error, VariablesOf<N, K>>({
    mutationFn: variables => (sdk[ns][method] as SdkMethod)(variables),
    ...options
  });
}

/**
 * @deprecated 请使用 `useSdkQuery` 直接绑定 sdk 方法
 */
export function createSdkQuery<TResult>(queryKey: string[], queryFn: () => Promise<TResult>) {
  return (options?: Omit<UseQueryOptions<TResult, Error>, "queryKey" | "queryFn">) => {
    return useQuery({
      queryKey,
      queryFn,
      ...options
    });
  };
}
