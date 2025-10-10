import { FormEvent, useEffect, useMemo, useState } from "react";
import "./App.css";
import { api } from "./api/client";
import {
  UsersQuery,
  TransactionsQuery,
  UsersResponse,
  TransactionsResponse,
  UserRelationshipsResponse,
  TransactionRelationshipsResponse
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
