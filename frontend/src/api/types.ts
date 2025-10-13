export interface Pagination {
  page: number;
  pageSize: number;
  totalItems: number;
  totalPages: number;
}

export interface UserSummary {
  userId: string;
  fullName: string;
  email: string;
  phone: string;
  kycStatus: string;
  riskScore: number;
  createdAt: string;
  updatedAt: string;
}

export interface TransactionSummary {
  transactionId: string;
  senderUserId: string;
  receiverUserId: string;
  amount: number;
  currency: string;
  type: string;
  status: string;
  channel: string;
  timestamp: string;
  createdAt: string;
  updatedAt: string;
}

export interface UsersResponse {
  items: UserSummary[];
  pagination: Pagination;
}

export interface TransactionsResponse {
  items: TransactionSummary[];
  pagination: Pagination;
}

export interface DirectConnection {
  userId: string;
  linkType: string;
  direction: string;
  transactionId: string;
  amount: number;
  currency: string;
  timestamp: string;
}

export interface UserTransactionLink {
  transactionId: string;
  role: string;
  amount: number;
  currency: string;
  timestamp: string;
}

export interface SharedAttribute {
  attributeType: string;
  attributeHash: string;
  connectedUsers: string[];
}

export interface UserRelationshipsResponse {
  userId: string;
  directConnections: DirectConnection[];
  transactions: UserTransactionLink[];
  sharedAttributes: SharedAttribute[];
}

export interface TransactionUserLink {
  userId: string;
  role: string;
  amount: number;
  currency: string;
  direction: string;
}

export interface LinkedTransaction {
  transactionId: string;
  linkType: string;
  attributeHash: string;
  score: number;
  updatedAt: string;
}

export interface TransactionRelationshipsResponse {
  transactionId: string;
  users: TransactionUserLink[];
  linkedTransactions: LinkedTransaction[];
}

export interface AddressInput {
  line1: string;
  line2?: string;
  city: string;
  state: string;
  postalCode: string;
  country: string;
}

export interface PaymentMethodInput {
  paymentMethodId: string;
  methodType: string;
  provider: string;
  masked?: string;
  fingerprint: string;
  firstUsedAt?: string;
  lastUsedAt?: string;
}

export interface CreateUserRequest {
  userId: string;
  fullName: string;
  email: string;
  phone: string;
  address: AddressInput;
  dateOfBirth?: string;
  kycStatus: string;
  riskScore: number;
  paymentMethods?: PaymentMethodInput[];
  attributes?: Array<Record<string, unknown>>;
  createdAt?: string;
  updatedAt?: string;
}

export interface CreateTransactionRequest {
  transactionId: string;
  senderUserId: string;
  receiverUserId: string;
  amount: number;
  currency: string;
  type: string;
  status: string;
  channel: string;
  ipAddress?: string;
  deviceId?: string;
  paymentMethodId?: string;
  timestamp: string;
  metadata?: Record<string, unknown>;
  createdAt?: string;
  updatedAt?: string;
}

export interface StatusResponse {
  status: string;
  id: string;
}

export interface UsersQuery {
  page: number;
  pageSize: number;
  search?: string;
  kycStatus?: string;
  riskMin?: number;
  riskMax?: number;
}

export interface TransactionsQuery {
  page: number;
  pageSize: number;
  search?: string;
  userId?: string;
  status?: string;
  type?: string;
  minAmount?: number;
  maxAmount?: number;
  start?: string;
  end?: string;
}
