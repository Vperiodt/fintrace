import {
  UsersQuery,
  TransactionsQuery,
  UsersResponse,
  TransactionsResponse,
  UserRelationshipsResponse,
  TransactionRelationshipsResponse,
  CreateUserRequest,
  CreateTransactionRequest,
  StatusResponse
} from "./types";

const rawApiBase = (import.meta.env.VITE_API_BASE_URL as string | undefined)?.trim();
// Hosts that only resolve inside the Docker Compose network.
const dockerOnlyHosts = new Set(["backend", "fintrace-backend"]);
const localHosts = new Set(["localhost", "127.0.0.1", "0.0.0.0"]);

const isAbsoluteUrl = (value: string) => /^https?:\/\//i.test(value);

const normalizeRelativeBase = (value: string) => {
  const trimmed = value.replace(/^\/+/, "").replace(/\/+$/, "");
  return trimmed ? `/${trimmed}` : "";
};

const shouldTreatAsRelative = (parsed: URL) => {
  if (dockerOnlyHosts.has(parsed.hostname) || localHosts.has(parsed.hostname)) {
    return true;
  }
  if (typeof window !== "undefined") {
    return parsed.hostname === window.location.hostname;
  }
  return false;
};

const resolveApiPrefix = () => {
  if (!rawApiBase) {
    return "";
  }

  if (!isAbsoluteUrl(rawApiBase)) {
    return normalizeRelativeBase(rawApiBase);
  }

  try {
    const parsed = new URL(rawApiBase);
    const path = parsed.pathname.replace(/\/+$/, "");
    if (shouldTreatAsRelative(parsed)) {
      return normalizeRelativeBase(path);
    }
    return `${parsed.origin}${path}`;
  } catch {
    return "";
  }
};

const API_PREFIX = resolveApiPrefix();

const toURL = (path: string) => {
  if (!API_PREFIX) {
    return path;
  }
  return `${API_PREFIX}${path}`;
};

async function fetchJSON<T>(input: RequestInfo | URL, init?: RequestInit): Promise<T> {
  const response = await fetch(input, init);
  if (!response.ok) {
    let message = response.statusText;
    try {
      const body = await response.json();
      if (body && typeof body.error === "string") {
        message = body.error;
      }
    } catch {
      // ignore body parsing errors
    }
    throw new Error(message || "Request failed");
  }
  return (await response.json()) as T;
}

const buildQuery = (params: Record<string, string | number | undefined | null>) => {
  const searchParams = new URLSearchParams();
  for (const [key, value] of Object.entries(params)) {
    if (value === undefined || value === null || value === "") {
      continue;
    }
    searchParams.set(key, String(value));
  }
  const queryString = searchParams.toString();
  return queryString ? `?${queryString}` : "";
};

export const api = {
  listUsers: async (query: UsersQuery): Promise<UsersResponse> => {
    const qs = buildQuery({
      page: query.page,
      pageSize: query.pageSize,
      search: query.search,
      kycStatus: query.kycStatus,
      riskMin: query.riskMin,
      riskMax: query.riskMax
    });
    return fetchJSON<UsersResponse>(toURL(`/users${qs}`));
  },

  listTransactions: async (query: TransactionsQuery): Promise<TransactionsResponse> => {
    const qs = buildQuery({
      page: query.page,
      pageSize: query.pageSize,
      search: query.search,
      userId: query.userId,
      status: query.status,
      type: query.type,
      minAmount: query.minAmount,
      maxAmount: query.maxAmount,
      start: query.start,
      end: query.end
    });
    return fetchJSON<TransactionsResponse>(toURL(`/transactions${qs}`));
  },

  getUserRelationships: async (userId: string): Promise<UserRelationshipsResponse> => {
    return fetchJSON<UserRelationshipsResponse>(
      toURL(`/relationships/user/${encodeURIComponent(userId)}`)
    );
  },

  getTransactionRelationships: async (
    transactionId: string
  ): Promise<TransactionRelationshipsResponse> => {
    return fetchJSON<TransactionRelationshipsResponse>(
      toURL(`/relationships/transaction/${encodeURIComponent(transactionId)}`)
    );
  },

  createUser: async (payload: CreateUserRequest): Promise<StatusResponse> => {
    return fetchJSON<StatusResponse>(toURL("/users"), {
      method: "POST",
      headers: {
        "Content-Type": "application/json"
      },
      body: JSON.stringify(payload)
    });
  },

  createTransaction: async (
    payload: CreateTransactionRequest
  ): Promise<StatusResponse> => {
    return fetchJSON<StatusResponse>(toURL("/transactions"), {
      method: "POST",
      headers: {
        "Content-Type": "application/json"
      },
      body: JSON.stringify(payload)
    });
  }
};

export type APIClient = typeof api;
