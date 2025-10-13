import { FormEvent, useEffect, useMemo, useState } from "react";
import "./App.css";
import { api } from "./api/client";
import {
  UsersQuery,
  TransactionsQuery,
  UsersResponse,
  TransactionsResponse,
  UserRelationshipsResponse,
  TransactionRelationshipsResponse,
  CreateUserRequest,
  CreateTransactionRequest
} from "./api/types";
import { UserTable } from "./components/UserTable";
import { TransactionTable } from "./components/TransactionTable";
import {
  GraphEdgeSelection,
  GraphExplorer,
  GraphNodeSelection
} from "./components/GraphExplorer";
import { GraphDetailsPanel } from "./components/GraphDetailsPanel";

const DEFAULT_USER_QUERY: UsersQuery = {
  page: 1,
  pageSize: 20
};

const DEFAULT_TRANSACTION_QUERY: TransactionsQuery = {
  page: 1,
  pageSize: 20
};

function App() {
  const [userQuery, setUserQuery] = useState<UsersQuery>(DEFAULT_USER_QUERY);
  const [transactionQuery, setTransactionQuery] = useState<TransactionsQuery>(
    DEFAULT_TRANSACTION_QUERY
  );

  const [userFilters, setUserFilters] = useState({
    search: "",
    kycStatus: "",
    riskMin: "",
    riskMax: ""
  });
  const [userFilterError, setUserFilterError] = useState<string | null>(null);

  const [transactionFilters, setTransactionFilters] = useState({
    search: "",
    userId: "",
    status: "",
    type: "",
    minAmount: "",
    maxAmount: "",
    start: "",
    end: ""
  });
  const [transactionFilterError, setTransactionFilterError] = useState<string | null>(
    null
  );

  const makeInitialUserForm = () => ({
    userId: "",
    fullName: "",
    email: "",
    phone: "",
    kycStatus: "VERIFIED",
    riskScore: "",
    addressLine1: "",
    addressLine2: "",
    addressCity: "",
    addressState: "",
    addressPostalCode: "",
    addressCountry: "US",
    paymentMethodId: "",
    paymentMethodProvider: "",
    paymentMethodMasked: "",
    paymentMethodFingerprint: ""
  });

  const makeInitialTransactionForm = () => ({
    transactionId: "",
    senderUserId: "",
    receiverUserId: "",
    amount: "",
    currency: "USD",
    type: "TRANSFER",
    status: "COMPLETED",
    channel: "WEB",
    timestamp: new Date().toISOString().slice(0, 16),
    ipAddress: "",
    deviceId: "",
    paymentMethodId: "",
    note: ""
  });

  const [userForm, setUserForm] = useState(makeInitialUserForm);
  const [userCreateError, setUserCreateError] = useState<string | null>(null);
  const [userCreateMessage, setUserCreateMessage] = useState<string | null>(null);

  const [transactionForm, setTransactionForm] = useState(makeInitialTransactionForm);
  const [transactionCreateError, setTransactionCreateError] = useState<string | null>(null);
  const [transactionCreateMessage, setTransactionCreateMessage] =
    useState<string | null>(null);

  const [usersData, setUsersData] = useState<UsersResponse | null>(null);
  const [usersLoading, setUsersLoading] = useState(false);
  const [usersError, setUsersError] = useState<string | null>(null);

  const [transactionsData, setTransactionsData] = useState<TransactionsResponse | null>(
    null
  );
  const [transactionsLoading, setTransactionsLoading] = useState(false);
  const [transactionsError, setTransactionsError] = useState<string | null>(null);

  const [selectedUserId, setSelectedUserId] = useState<string | null>(null);
  const [selectedTransactionId, setSelectedTransactionId] = useState<string | null>(
    null
  );

  const [userRelationships, setUserRelationships] =
    useState<UserRelationshipsResponse | null>(null);
  const [userRelationshipsLoading, setUserRelationshipsLoading] = useState(false);
  const [userRelationshipsError, setUserRelationshipsError] = useState<string | null>(
    null
  );

  const [transactionRelationships, setTransactionRelationships] =
    useState<TransactionRelationshipsResponse | null>(null);
  const [transactionRelationshipsLoading, setTransactionRelationshipsLoading] =
    useState(false);
  const [transactionRelationshipsError, setTransactionRelationshipsError] =
    useState<string | null>(null);

  const [graphSelection, setGraphSelection] = useState<GraphNodeSelection | null>(
    null
  );
  const [graphEdgeSelection, setGraphEdgeSelection] =
    useState<GraphEdgeSelection | null>(null);
  const [graphSearchTerm, setGraphSearchTerm] = useState("");

  const usersList = usersData?.items ?? [];
  const transactionsList = transactionsData?.items ?? [];

  const selectedUserSummary = useMemo(() => {
    if (!selectedUserId) {
      return null;
    }
    return usersList.find((user) => user.userId === selectedUserId) ?? null;
  }, [usersList, selectedUserId]);

  const selectedTransactionSummary = useMemo(() => {
    if (!selectedTransactionId) {
      return null;
    }
    return (
      transactionsList.find((tx) => tx.transactionId === selectedTransactionId) ??
      null
    );
  }, [transactionsList, selectedTransactionId]);

  useEffect(() => {
    let ignore = false;
    setUsersLoading(true);
    api
      .listUsers(userQuery)
      .then((data) => {
        if (ignore) return;
        setUsersData(data);
        setUsersError(null);
      })
      .catch((err) => {
        if (ignore) return;
        setUsersError(err.message);
      })
      .finally(() => {
        if (!ignore) {
          setUsersLoading(false);
        }
      });
    return () => {
      ignore = true;
    };
  }, [userQuery]);

  useEffect(() => {
    let ignore = false;
    setTransactionsLoading(true);
    api
      .listTransactions(transactionQuery)
      .then((data) => {
        if (ignore) return;
        setTransactionsData(data);
        setTransactionsError(null);
      })
      .catch((err) => {
        if (ignore) return;
        setTransactionsError(err.message);
      })
      .finally(() => {
        if (!ignore) {
          setTransactionsLoading(false);
        }
      });
    return () => {
      ignore = true;
    };
  }, [transactionQuery]);

  useEffect(() => {
    if (!selectedUserId) {
      setUserRelationships(null);
      setUserRelationshipsError(null);
      setUserRelationshipsLoading(false);
      return;
    }
    let ignore = false;
    setUserRelationshipsLoading(true);
    api
      .getUserRelationships(selectedUserId)
      .then((data) => {
        if (ignore) return;
        setUserRelationships(data);
        setUserRelationshipsError(null);
      })
      .catch((err) => {
        if (ignore) return;
        setUserRelationshipsError(err.message);
      })
      .finally(() => {
        if (!ignore) {
          setUserRelationshipsLoading(false);
        }
      });
    return () => {
      ignore = true;
    };
  }, [selectedUserId]);

  useEffect(() => {
    if (!selectedTransactionId) {
      setTransactionRelationships(null);
      setTransactionRelationshipsError(null);
      setTransactionRelationshipsLoading(false);
      return;
    }
    let ignore = false;
    setTransactionRelationshipsLoading(true);
    api
      .getTransactionRelationships(selectedTransactionId)
      .then((data) => {
        if (ignore) return;
        setTransactionRelationships(data);
        setTransactionRelationshipsError(null);
      })
      .catch((err) => {
        if (ignore) return;
        setTransactionRelationshipsError(err.message);
      })
      .finally(() => {
        if (!ignore) {
          setTransactionRelationshipsLoading(false);
        }
      });
    return () => {
      ignore = true;
    };
  }, [selectedTransactionId]);

  const activeMode = useMemo<"user" | "transaction" | null>(() => {
    if (selectedUserId) return "user";
    if (selectedTransactionId) return "transaction";
    return null;
  }, [selectedUserId, selectedTransactionId]);

  const graphLoading = activeMode === "user"
    ? userRelationshipsLoading
    : activeMode === "transaction"
    ? transactionRelationshipsLoading
    : false;

  const graphError = activeMode === "user"
    ? userRelationshipsError
    : activeMode === "transaction"
    ? transactionRelationshipsError
    : null;

  useEffect(() => {
    setGraphSelection(null);
    setGraphEdgeSelection(null);
  }, [activeMode, selectedUserId, selectedTransactionId]);

  const handleApplyUserFilters = (event: FormEvent) => {
    event.preventDefault();

    const { search, kycStatus, riskMin, riskMax } = userFilters;
    const nextQuery: UsersQuery = {
      ...userQuery,
      page: 1,
      search: search || undefined,
      kycStatus: kycStatus || undefined,
      pageSize: userQuery.pageSize
    };

    if (riskMin) {
      const parsed = parseFloat(riskMin);
      if (Number.isNaN(parsed)) {
        setUserFilterError("riskMin must be a valid number between 0 and 1.");
        return;
      }
      nextQuery.riskMin = parsed;
    } else {
      delete nextQuery.riskMin;
    }

    if (riskMax) {
      const parsed = parseFloat(riskMax);
      if (Number.isNaN(parsed)) {
        setUserFilterError("riskMax must be a valid number between 0 and 1.");
        return;
      }
      nextQuery.riskMax = parsed;
    } else {
      delete nextQuery.riskMax;
    }

    setUserFilterError(null);
    setSelectedUserId(null);
    setSelectedTransactionId(null);
    setUserQuery(nextQuery);
  };

  const handleApplyTransactionFilters = (event: FormEvent) => {
    event.preventDefault();

    const { search, userId, status, type, minAmount, maxAmount, start, end } =
      transactionFilters;
    const nextQuery: TransactionsQuery = {
      ...transactionQuery,
      page: 1,
      search: search || undefined,
      userId: userId || undefined,
      status: status || undefined,
      type: type || undefined,
      pageSize: transactionQuery.pageSize
    };

    if (minAmount) {
      const parsed = parseFloat(minAmount);
      if (Number.isNaN(parsed)) {
        setTransactionFilterError("minAmount must be a valid number.");
        return;
      }
      nextQuery.minAmount = parsed;
    } else {
      delete nextQuery.minAmount;
    }

    if (maxAmount) {
      const parsed = parseFloat(maxAmount);
      if (Number.isNaN(parsed)) {
        setTransactionFilterError("maxAmount must be a valid number.");
        return;
      }
      nextQuery.maxAmount = parsed;
    } else {
      delete nextQuery.maxAmount;
    }

    if (start) {
      const date = new Date(start);
      if (Number.isNaN(date.getTime())) {
        setTransactionFilterError("start must be a valid datetime.");
        return;
      }
      nextQuery.start = date.toISOString();
    } else {
      delete nextQuery.start;
    }

    if (end) {
      const date = new Date(end);
      if (Number.isNaN(date.getTime())) {
        setTransactionFilterError("end must be a valid datetime.");
        return;
      }
      nextQuery.end = date.toISOString();
    } else {
      delete nextQuery.end;
    }

    setTransactionFilterError(null);
    setSelectedTransactionId(null);
    setSelectedUserId(null);
    setTransactionQuery(nextQuery);
  };

  const handleUserPageChange = (page: number) => {
    if (page < 1 || (usersData?.pagination.totalPages ?? 0) === 0) {
      return;
    }
    const totalPages = usersData?.pagination.totalPages ?? 1;
    const nextPage = Math.max(1, Math.min(page, totalPages));
    setUserQuery((prev) => ({ ...prev, page: nextPage }));
  };

  const handleTransactionPageChange = (page: number) => {
    if (page < 1 || (transactionsData?.pagination.totalPages ?? 0) === 0) {
      return;
    }
    const totalPages = transactionsData?.pagination.totalPages ?? 1;
    const nextPage = Math.max(1, Math.min(page, totalPages));
    setTransactionQuery((prev) => ({ ...prev, page: nextPage }));
  };

  const handleUserFormSubmit = async (event: FormEvent) => {
    event.preventDefault();
    setUserCreateError(null);
    setUserCreateMessage(null);

    const trimmedId = userForm.userId.trim();
    if (!trimmedId) {
      setUserCreateError("User ID is required.");
      return;
    }
    if (!userForm.fullName.trim()) {
      setUserCreateError("Full name is required.");
      return;
    }
    const risk = Number(userForm.riskScore);
    if (Number.isNaN(risk) || risk < 0 || risk > 1) {
      setUserCreateError("Risk score must be between 0 and 1.");
      return;
    }

    const nowIso = new Date().toISOString();
    const payload = {
      userId: trimmedId,
      fullName: userForm.fullName.trim(),
      email: userForm.email.trim(),
      phone: userForm.phone.trim(),
      address: {
        line1: userForm.addressLine1.trim(),
        line2: userForm.addressLine2.trim(),
        city: userForm.addressCity.trim(),
        state: userForm.addressState.trim(),
        postalCode: userForm.addressPostalCode.trim(),
        country: userForm.addressCountry.trim() || "US"
      },
      kycStatus: userForm.kycStatus,
      riskScore: risk,
      createdAt: nowIso,
      updatedAt: nowIso
    } as CreateUserRequest;

    const paymentFingerprint = userForm.paymentMethodFingerprint.trim();
    const paymentId = userForm.paymentMethodId.trim();
    if (paymentFingerprint && paymentId) {
      payload.paymentMethods = [
        {
          paymentMethodId: paymentId,
          methodType: "CARD",
          provider: userForm.paymentMethodProvider.trim() || "CARD",
          masked: userForm.paymentMethodMasked.trim() || undefined,
          fingerprint: paymentFingerprint,
          firstUsedAt: nowIso,
          lastUsedAt: nowIso
        }
      ];
    }

    try {
      await api.createUser(payload);
      setUserCreateMessage(`User ${payload.userId} saved.`);
      setUserForm(makeInitialUserForm());
      setUserQuery((prev) => ({ ...prev }));
    } catch (err) {
      setUserCreateError((err as Error).message);
    }
  };

  const handleTransactionFormSubmit = async (event: FormEvent) => {
    event.preventDefault();
    setTransactionCreateError(null);
    setTransactionCreateMessage(null);

    const trimmedId = transactionForm.transactionId.trim();
    if (!trimmedId) {
      setTransactionCreateError("Transaction ID is required.");
      return;
    }
    if (!transactionForm.senderUserId.trim() || !transactionForm.receiverUserId.trim()) {
      setTransactionCreateError("Sender and receiver user IDs are required.");
      return;
    }
    const amount = Number(transactionForm.amount);
    if (Number.isNaN(amount) || amount <= 0) {
      setTransactionCreateError("Amount must be a positive number.");
      return;
    }
    const timestampMs = Date.parse(transactionForm.timestamp);
    if (Number.isNaN(timestampMs)) {
      setTransactionCreateError("Timestamp must be a valid date/time.");
      return;
    }

    const now = new Date().toISOString();
    const payload = {
      transactionId: trimmedId,
      senderUserId: transactionForm.senderUserId.trim(),
      receiverUserId: transactionForm.receiverUserId.trim(),
      amount,
      currency: transactionForm.currency.trim() || "USD",
      type: transactionForm.type.trim() || "TRANSFER",
      status: transactionForm.status.trim() || "COMPLETED",
      channel: transactionForm.channel.trim() || "WEB",
      ipAddress: transactionForm.ipAddress.trim() || undefined,
      deviceId: transactionForm.deviceId.trim() || undefined,
      paymentMethodId: transactionForm.paymentMethodId.trim() || undefined,
      timestamp: new Date(timestampMs).toISOString(),
      metadata: transactionForm.note.trim() ? { note: transactionForm.note.trim() } : undefined,
      createdAt: now,
      updatedAt: now
    } as CreateTransactionRequest;

    try {
      await api.createTransaction(payload);
      setTransactionCreateMessage(`Transaction ${payload.transactionId} saved.`);
      setTransactionForm(makeInitialTransactionForm());
      setTransactionQuery((prev) => ({ ...prev }));
    } catch (err) {
      setTransactionCreateError((err as Error).message);
    }
  };

  const handleUserFormReset = () => {
    setUserForm(makeInitialUserForm());
    setUserCreateError(null);
    setUserCreateMessage(null);
  };

  const handleTransactionFormReset = () => {
    setTransactionForm(makeInitialTransactionForm());
    setTransactionCreateError(null);
    setTransactionCreateMessage(null);
  };

  const handleSelectUser = (userId: string) => {
    setSelectedUserId(userId);
    setSelectedTransactionId(null);
  };

  const handleSelectTransaction = (transactionId: string) => {
    setSelectedTransactionId(transactionId);
    setSelectedUserId(null);
  };

  return (
    <div className="app">
      <header>
        <h1>Fintrace Relationship Explorer</h1>
        <p>
          Visualise user and transaction relationships across direct transfers and
          shared attributes.
        </p>
      </header>

      <section className="panel create-panel">
        <h2>Create Data</h2>
        <p className="create-panel-description">
          Quickly add demo users and transactions without leaving the browser. Newly
          created entities appear in the tables below and wire into the graph
          automatically.
        </p>
        <div className="create-forms">
          <form className="create-form" onSubmit={handleUserFormSubmit}>
            <h3>New User</h3>
            <div className="form-grid">
              <label>
                User ID*
                <input
                  type="text"
                  value={userForm.userId}
                  onChange={(event) =>
                    setUserForm((prev) => ({ ...prev, userId: event.target.value }))
                  }
                  required
                />
              </label>
              <label>
                Full name*
                <input
                  type="text"
                  value={userForm.fullName}
                  onChange={(event) =>
                    setUserForm((prev) => ({ ...prev, fullName: event.target.value }))
                  }
                  required
                />
              </label>
              <label>
                Email
                <input
                  type="email"
                  value={userForm.email}
                  onChange={(event) =>
                    setUserForm((prev) => ({ ...prev, email: event.target.value }))
                  }
                />
              </label>
              <label>
                Phone
                <input
                  type="text"
                  value={userForm.phone}
                  onChange={(event) =>
                    setUserForm((prev) => ({ ...prev, phone: event.target.value }))
                  }
                />
              </label>
              <label>
                KYC status
                <select
                  value={userForm.kycStatus}
                  onChange={(event) =>
                    setUserForm((prev) => ({ ...prev, kycStatus: event.target.value }))
                  }
                >
                  <option value="VERIFIED">Verified</option>
                  <option value="PENDING">Pending</option>
                  <option value="REVIEW">Review</option>
                </select>
              </label>
              <label>
                Risk score (0-1)*
                <input
                  type="number"
                  min="0"
                  max="1"
                  step="0.01"
                  value={userForm.riskScore}
                  onChange={(event) =>
                    setUserForm((prev) => ({ ...prev, riskScore: event.target.value }))
                  }
                  placeholder="0.35"
                  required
                />
              </label>
            </div>

            <h4>Address</h4>
            <div className="form-grid">
              <label>
                Line 1
                <input
                  type="text"
                  value={userForm.addressLine1}
                  onChange={(event) =>
                    setUserForm((prev) => ({ ...prev, addressLine1: event.target.value }))
                  }
                />
              </label>
              <label>
                Line 2
                <input
                  type="text"
                  value={userForm.addressLine2}
                  onChange={(event) =>
                    setUserForm((prev) => ({ ...prev, addressLine2: event.target.value }))
                  }
                />
              </label>
              <label>
                City
                <input
                  type="text"
                  value={userForm.addressCity}
                  onChange={(event) =>
                    setUserForm((prev) => ({ ...prev, addressCity: event.target.value }))
                  }
                />
              </label>
              <label>
                State
                <input
                  type="text"
                  value={userForm.addressState}
                  onChange={(event) =>
                    setUserForm((prev) => ({ ...prev, addressState: event.target.value }))
                  }
                />
              </label>
              <label>
                Postal code
                <input
                  type="text"
                  value={userForm.addressPostalCode}
                  onChange={(event) =>
                    setUserForm((prev) => ({
                      ...prev,
                      addressPostalCode: event.target.value
                    }))
                  }
                />
              </label>
              <label>
                Country
                <input
                  type="text"
                  value={userForm.addressCountry}
                  onChange={(event) =>
                    setUserForm((prev) => ({
                      ...prev,
                      addressCountry: event.target.value
                    }))
                  }
                />
              </label>
            </div>

            <details className="optional-block">
              <summary>Optional payment method</summary>
              <div className="form-grid">
                <label>
                  Payment method ID
                  <input
                    type="text"
                    value={userForm.paymentMethodId}
                    onChange={(event) =>
                      setUserForm((prev) => ({
                        ...prev,
                        paymentMethodId: event.target.value
                      }))
                    }
                  />
                </label>
                <label>
                  Provider
                  <input
                    type="text"
                    value={userForm.paymentMethodProvider}
                    onChange={(event) =>
                      setUserForm((prev) => ({
                        ...prev,
                        paymentMethodProvider: event.target.value
                      }))
                    }
                  />
                </label>
                <label>
                  Masked PAN
                  <input
                    type="text"
                    value={userForm.paymentMethodMasked}
                    onChange={(event) =>
                      setUserForm((prev) => ({
                        ...prev,
                        paymentMethodMasked: event.target.value
                      }))
                    }
                  />
                </label>
                <label>
                  Fingerprint
                  <input
                    type="text"
                    value={userForm.paymentMethodFingerprint}
                    onChange={(event) =>
                      setUserForm((prev) => ({
                        ...prev,
                        paymentMethodFingerprint: event.target.value
                      }))
                    }
                  />
                </label>
              </div>
            </details>

            {userCreateError && <div className="error">{userCreateError}</div>}
            {userCreateMessage && <div className="success">{userCreateMessage}</div>}

            <div className="form-actions">
              <button type="submit">Save user</button>
              <button type="button" className="secondary" onClick={handleUserFormReset}>
                Clear
              </button>
            </div>
          </form>

          <form className="create-form" onSubmit={handleTransactionFormSubmit}>
            <h3>New Transaction</h3>
            <div className="form-grid">
              <label>
                Transaction ID*
                <input
                  type="text"
                  value={transactionForm.transactionId}
                  onChange={(event) =>
                    setTransactionForm((prev) => ({
                      ...prev,
                      transactionId: event.target.value
                    }))
                  }
                  required
                />
              </label>
              <label>
                Sender user ID*
                <input
                  type="text"
                  value={transactionForm.senderUserId}
                  onChange={(event) =>
                    setTransactionForm((prev) => ({
                      ...prev,
                      senderUserId: event.target.value
                    }))
                  }
                  required
                />
              </label>
              <label>
                Receiver user ID*
                <input
                  type="text"
                  value={transactionForm.receiverUserId}
                  onChange={(event) =>
                    setTransactionForm((prev) => ({
                      ...prev,
                      receiverUserId: event.target.value
                    }))
                  }
                  required
                />
              </label>
              <label>
                Amount (USD)*
                <input
                  type="number"
                  min="0"
                  step="0.01"
                  value={transactionForm.amount}
                  onChange={(event) =>
                    setTransactionForm((prev) => ({
                      ...prev,
                      amount: event.target.value
                    }))
                  }
                  required
                />
              </label>
              <label>
                Currency
                <input
                  type="text"
                  value={transactionForm.currency}
                  onChange={(event) =>
                    setTransactionForm((prev) => ({
                      ...prev,
                      currency: event.target.value
                    }))
                  }
                />
              </label>
              <label>
                Type
                <select
                  value={transactionForm.type}
                  onChange={(event) =>
                    setTransactionForm((prev) => ({ ...prev, type: event.target.value }))
                  }
                >
                  <option value="TRANSFER">Transfer</option>
                  <option value="PAYMENT">Payment</option>
                  <option value="REFUND">Refund</option>
                  <option value="INVOICE">Invoice</option>
                </select>
              </label>
              <label>
                Status
                <select
                  value={transactionForm.status}
                  onChange={(event) =>
                    setTransactionForm((prev) => ({
                      ...prev,
                      status: event.target.value
                    }))
                  }
                >
                  <option value="COMPLETED">Completed</option>
                  <option value="PENDING">Pending</option>
                  <option value="FAILED">Failed</option>
                </select>
              </label>
              <label>
                Channel
                <input
                  type="text"
                  value={transactionForm.channel}
                  onChange={(event) =>
                    setTransactionForm((prev) => ({
                      ...prev,
                      channel: event.target.value
                    }))
                  }
                />
              </label>
              <label>
                Timestamp*
                <input
                  type="datetime-local"
                  value={transactionForm.timestamp}
                  onChange={(event) =>
                    setTransactionForm((prev) => ({
                      ...prev,
                      timestamp: event.target.value
                    }))
                  }
                  required
                />
              </label>
            </div>

            <div className="form-grid">
              <label>
                IP address
                <input
                  type="text"
                  value={transactionForm.ipAddress}
                  onChange={(event) =>
                    setTransactionForm((prev) => ({
                      ...prev,
                      ipAddress: event.target.value
                    }))
                  }
                />
              </label>
              <label>
                Device ID
                <input
                  type="text"
                  value={transactionForm.deviceId}
                  onChange={(event) =>
                    setTransactionForm((prev) => ({
                      ...prev,
                      deviceId: event.target.value
                    }))
                  }
                />
              </label>
              <label>
                Payment method ID
                <input
                  type="text"
                  value={transactionForm.paymentMethodId}
                  onChange={(event) =>
                    setTransactionForm((prev) => ({
                      ...prev,
                      paymentMethodId: event.target.value
                    }))
                  }
                />
              </label>
              <label>
                Note
                <input
                  type="text"
                  value={transactionForm.note}
                  onChange={(event) =>
                    setTransactionForm((prev) => ({ ...prev, note: event.target.value }))
                  }
                />
              </label>
            </div>

            {transactionCreateError && <div className="error">{transactionCreateError}</div>}
            {transactionCreateMessage && (
              <div className="success">{transactionCreateMessage}</div>
            )}

            <div className="form-actions">
              <button type="submit">Save transaction</button>
              <button
                type="button"
                className="secondary"
                onClick={handleTransactionFormReset}
              >
                Clear
              </button>
            </div>
          </form>
        </div>
      </section>

      <main className="layout">
        <section className="panel">
          <div className="panel-header">
            <h2>Users</h2>
            <div className="page-size-selector">
              <label>
                Page size
                <select
                  value={userQuery.pageSize}
                  onChange={(event) =>
                    setUserQuery((prev) => ({
                      ...prev,
                      pageSize: Number(event.target.value),
                      page: 1
                    }))
                  }
                >
                  <option value={20}>20</option>
                  <option value={50}>50</option>
                  <option value={100}>100</option>
                </select>
              </label>
            </div>
          </div>
          <form className="filter-form" onSubmit={handleApplyUserFilters}>
            <input
              type="text"
              placeholder="Search name, email, or user ID"
              value={userFilters.search}
              onChange={(event) =>
                setUserFilters((prev) => ({ ...prev, search: event.target.value }))
              }
            />
            <select
              value={userFilters.kycStatus}
              onChange={(event) =>
                setUserFilters((prev) => ({ ...prev, kycStatus: event.target.value }))
              }
            >
              <option value="">All KYC statuses</option>
              <option value="VERIFIED">Verified</option>
              <option value="PENDING">Pending</option>
              <option value="REVIEW">Review</option>
            </select>
            <input
              type="number"
              step="0.01"
              min="0"
              max="1"
              placeholder="Risk min"
              value={userFilters.riskMin}
              onChange={(event) =>
                setUserFilters((prev) => ({ ...prev, riskMin: event.target.value }))
              }
            />
            <input
              type="number"
              step="0.01"
              min="0"
              max="1"
              placeholder="Risk max"
              value={userFilters.riskMax}
              onChange={(event) =>
                setUserFilters((prev) => ({ ...prev, riskMax: event.target.value }))
              }
            />
            <button type="submit">Apply</button>
            <button
              type="button"
              className="secondary"
              onClick={() => {
                setUserFilters({ search: "", kycStatus: "", riskMin: "", riskMax: "" });
                setUserFilterError(null);
                setUserQuery(DEFAULT_USER_QUERY);
                setSelectedUserId(null);
              }}
            >
              Reset
            </button>
          </form>
          {userFilterError && <div className="error">{userFilterError}</div>}
          <UserTable
            users={usersData?.items || []}
            loading={usersLoading}
            error={usersError}
            pagination={usersData?.pagination}
            onPageChange={handleUserPageChange}
            onSelect={handleSelectUser}
            selectedUserId={selectedUserId}
          />
        </section>

        <section className="panel">
          <div className="panel-header">
            <h2>Transactions</h2>
            <div className="page-size-selector">
              <label>
                Page size
                <select
                  value={transactionQuery.pageSize}
                  onChange={(event) =>
                    setTransactionQuery((prev) => ({
                      ...prev,
                      pageSize: Number(event.target.value),
                      page: 1
                    }))
                  }
                >
                  <option value={20}>20</option>
                  <option value={50}>50</option>
                  <option value={100}>100</option>
                </select>
              </label>
            </div>
          </div>
          <form className="filter-form" onSubmit={handleApplyTransactionFilters}>
            <input
              type="text"
              placeholder="Search transaction ID"
              value={transactionFilters.search}
              onChange={(event) =>
                setTransactionFilters((prev) => ({
                  ...prev,
                  search: event.target.value
                }))
              }
            />
            <input
              type="text"
              placeholder="User ID"
              value={transactionFilters.userId}
              onChange={(event) =>
                setTransactionFilters((prev) => ({
                  ...prev,
                  userId: event.target.value
                }))
              }
            />
            <select
              value={transactionFilters.status}
              onChange={(event) =>
                setTransactionFilters((prev) => ({
                  ...prev,
                  status: event.target.value
                }))
              }
            >
              <option value="">All statuses</option>
              <option value="COMPLETED">Completed</option>
              <option value="FAILED">Failed</option>
              <option value="PENDING">Pending</option>
            </select>
            <select
              value={transactionFilters.type}
              onChange={(event) =>
                setTransactionFilters((prev) => ({
                  ...prev,
                  type: event.target.value
                }))
              }
            >
              <option value="">All types</option>
              <option value="TRANSFER">Transfer</option>
              <option value="PAYMENT">Payment</option>
              <option value="WITHDRAWAL">Withdrawal</option>
              <option value="DEPOSIT">Deposit</option>
            </select>
            <input
              type="number"
              step="0.01"
              placeholder="Min amount"
              value={transactionFilters.minAmount}
              onChange={(event) =>
                setTransactionFilters((prev) => ({
                  ...prev,
                  minAmount: event.target.value
                }))
              }
            />
            <input
              type="number"
              step="0.01"
              placeholder="Max amount"
              value={transactionFilters.maxAmount}
              onChange={(event) =>
                setTransactionFilters((prev) => ({
                  ...prev,
                  maxAmount: event.target.value
                }))
              }
            />
            <input
              type="datetime-local"
              value={transactionFilters.start}
              onChange={(event) =>
                setTransactionFilters((prev) => ({
                  ...prev,
                  start: event.target.value
                }))
              }
            />
            <input
              type="datetime-local"
              value={transactionFilters.end}
              onChange={(event) =>
                setTransactionFilters((prev) => ({
                  ...prev,
                  end: event.target.value
                }))
              }
            />
            <button type="submit">Apply</button>
            <button
              type="button"
              className="secondary"
              onClick={() => {
                setTransactionFilters({
                  search: "",
                  userId: "",
                  status: "",
                  type: "",
                  minAmount: "",
                  maxAmount: "",
                  start: "",
                  end: ""
                });
                setTransactionFilterError(null);
                setTransactionQuery(DEFAULT_TRANSACTION_QUERY);
                setSelectedTransactionId(null);
              }}
            >
              Reset
            </button>
          </form>
          {transactionFilterError && <div className="error">{transactionFilterError}</div>}
          <TransactionTable
            transactions={transactionsData?.items || []}
            loading={transactionsLoading}
            error={transactionsError}
            pagination={transactionsData?.pagination}
            onPageChange={handleTransactionPageChange}
            onSelect={handleSelectTransaction}
            selectedTransactionId={selectedTransactionId}
          />
        </section>

        <section className="panel graph-panel">
          <div className="panel-header">
            <h2>Graph Explorer</h2>
            {activeMode === "user" && selectedUserId && (
              <span className="chip">User: {selectedUserId}</span>
            )}
            {activeMode === "transaction" && selectedTransactionId && (
              <span className="chip">Transaction: {selectedTransactionId}</span>
            )}
          </div>
          <div className="graph-search">
            <input
              type="text"
              value={graphSearchTerm}
              placeholder="Search within graph (user, transaction, attribute)…"
              onChange={(event) => setGraphSearchTerm(event.target.value)}
            />
            {graphSearchTerm && (
              <button
                type="button"
                onClick={() => setGraphSearchTerm("")}
                aria-label="Clear graph search"
              >
                Clear
              </button>
            )}
          </div>
          <div className="graph-layout">
            <div className="graph-main">
              <GraphExplorer
                mode={activeMode}
                userRelationships={activeMode === "user" ? userRelationships : null}
                transactionRelationships={
                  activeMode === "transaction" ? transactionRelationships : null
                }
                loading={graphLoading}
                error={graphError}
                onNodeSelect={setGraphSelection}
                onEdgeSelect={setGraphEdgeSelection}
                searchTerm={graphSearchTerm}
              />
            </div>
            <GraphDetailsPanel
              mode={activeMode}
              userSummary={selectedUserSummary}
              userRelationships={activeMode === "user" ? userRelationships : null}
              transactionSummary={selectedTransactionSummary}
              transactionRelationships={
                activeMode === "transaction" ? transactionRelationships : null
              }
              selectedNode={graphSelection}
              selectedEdge={graphEdgeSelection}
              allUsers={usersList}
              allTransactions={transactionsList}
            />
          </div>
        </section>
      </main>
    </div>
  );
}

export default App;
