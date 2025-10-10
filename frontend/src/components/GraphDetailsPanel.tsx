import { useMemo } from "react";
import {
  UserRelationshipsResponse,
  TransactionRelationshipsResponse,
  UserSummary,
  TransactionSummary
} from "../api/types";
import { GraphEdgeSelection, GraphNodeSelection } from "./GraphExplorer";
import { describeRiskScore, formatCurrency, formatDate, truncateId } from "../utils/format";

interface GraphDetailsPanelProps {
  mode: "user" | "transaction" | null;
  userSummary?: UserSummary | null;
  transactionSummary?: TransactionSummary | null;
  userRelationships?: UserRelationshipsResponse | null;
  transactionRelationships?: TransactionRelationshipsResponse | null;
  selectedNode?: GraphNodeSelection | null;
  selectedEdge?: GraphEdgeSelection | null;
  allUsers?: UserSummary[];
  allTransactions?: TransactionSummary[];
}

const safeParseTime = (value?: string): number => {
  if (!value) {
    return 0;
  }
  const parsed = Date.parse(value);
  return Number.isNaN(parsed) ? 0 : parsed;
};

const prettifyAttributeType = (value: string): string =>
  value
    .toLowerCase()
    .split("_")
    .map((chunk) => chunk.charAt(0).toUpperCase() + chunk.slice(1))
    .join(" ");

export function GraphDetailsPanel({
  mode,
  userSummary,
  transactionSummary,
  userRelationships,
  transactionRelationships,
  selectedNode,
  selectedEdge,
  allUsers = [],
  allTransactions = []
}: GraphDetailsPanelProps) {
  const sortedUserTransactions = useMemo(() => {
    if (!userRelationships) {
      return [];
    }
    return [...userRelationships.transactions].sort(
      (a, b) => safeParseTime(b.timestamp) - safeParseTime(a.timestamp)
    );
  }, [userRelationships]);

  const sortedLinkedTransactions = useMemo(() => {
    if (!transactionRelationships) {
      return [];
    }
    return [...transactionRelationships.linkedTransactions].sort(
      (a, b) => safeParseTime(b.updatedAt) - safeParseTime(a.updatedAt)
    );
  }, [transactionRelationships]);

  const renderPrimaryCard = () => {
    if (mode === "user") {
      if (!userSummary) {
        return <p className="details-placeholder">Select a user to view profile insights.</p>;
      }

      const latestTx = sortedUserTransactions[0];

      return (
        <>
          <h3>User Details</h3>
          <p className="details-heading">{userSummary.userId}</p>
          <ul className="details-list">
            <li>
              <span className="label">Name</span>
              <span className="value">{userSummary.fullName || "—"}</span>
            </li>
            <li>
              <span className="label">Email</span>
              <span className="value monospace">{userSummary.email || "—"}</span>
            </li>
            <li>
              <span className="label">Phone</span>
              <span className="value">{userSummary.phone || "—"}</span>
            </li>
            <li>
              <span className="label">KYC Status</span>
              <span className="value">{userSummary.kycStatus || "—"}</span>
            </li>
            <li>
              <span className="label">Risk Score</span>
              <span className="value">{describeRiskScore(userSummary.riskScore)}</span>
            </li>
          </ul>

          {latestTx && (
            <div className="details-section">
              <h4 className="details-subheading">Most Recent Transaction</h4>
              <ul className="details-list compact">
                <li>
                  <span className="label">Transaction ID</span>
                  <span className="value monospace">{truncateId(latestTx.transactionId)}</span>
                </li>
                <li>
                  <span className="label">Role</span>
                  <span className="value">{latestTx.role}</span>
                </li>
                <li>
                  <span className="label">Amount</span>
                  <span className="value">
                    {formatCurrency(latestTx.amount, latestTx.currency)}
                  </span>
                </li>
                <li>
                  <span className="label">Date</span>
                  <span className="value">{formatDate(latestTx.timestamp)}</span>
                </li>
              </ul>
            </div>
          )}

          {sortedUserTransactions.length > 1 && (
            <div className="details-section">
              <h4 className="details-subheading">Linked Transactions</h4>
              <ul className="details-linked">
                {sortedUserTransactions.slice(1, 5).map((tx) => (
                  <li key={tx.transactionId}>
                    <span className="title monospace">{truncateId(tx.transactionId)}</span>
                    <span className="meta">
                      {tx.role} · {formatCurrency(tx.amount, tx.currency)} ·{" "}
                      {formatDate(tx.timestamp)}
                    </span>
                  </li>
                ))}
              </ul>
            </div>
          )}
        </>
      );
    }

    if (mode === "transaction") {
      if (!transactionSummary) {
        return (
          <p className="details-placeholder">Select a transaction to see participant context.</p>
        );
      }

      return (
        <>
          <h3>Transaction Details</h3>
          <p className="details-heading">{transactionSummary.transactionId}</p>
          <ul className="details-list">
            <li>
              <span className="label">Amount</span>
              <span className="value">
                {formatCurrency(transactionSummary.amount, transactionSummary.currency)}
              </span>
            </li>
            <li>
              <span className="label">Status</span>
              <span className="value">{transactionSummary.status}</span>
            </li>
            <li>
              <span className="label">Type</span>
              <span className="value">{transactionSummary.type}</span>
            </li>
            <li>
              <span className="label">Channel</span>
              <span className="value">{transactionSummary.channel}</span>
            </li>
            <li>
              <span className="label">Timestamp</span>
              <span className="value">{formatDate(transactionSummary.timestamp)}</span>
            </li>
          </ul>

          {transactionRelationships && transactionRelationships.users.length > 0 && (
            <div className="details-section">
              <h4 className="details-subheading">Participants</h4>
              <ul className="details-linked">
                {transactionRelationships.users.map((participant) => (
                  <li key={`${participant.userId}-${participant.role}`}>
                    <span className="title monospace">{truncateId(participant.userId)}</span>
                    <span className="meta">
                      {participant.role} · {participant.direction} ·{" "}
                      {formatCurrency(participant.amount, participant.currency)}
                    </span>
                  </li>
                ))}
              </ul>
            </div>
          )}

          {sortedLinkedTransactions.length > 0 && (
            <div className="details-section">
              <h4 className="details-subheading">Linked Transactions</h4>
              <ul className="details-linked">
                {sortedLinkedTransactions.slice(0, 4).map((link) => (
                  <li key={link.transactionId}>
                    <span className="title monospace">{truncateId(link.transactionId)}</span>
                    <span className="meta">
                      {link.linkType} · score {link.score.toFixed(2)} ·{" "}
                      {formatDate(link.updatedAt)}
                    </span>
                  </li>
                ))}
              </ul>
            </div>
          )}
        </>
      );
    }

    return (
      <p className="details-placeholder">
        Select a user or transaction from the tables to inspect relationships.
      </p>
    );
  };

  const renderNodeDetails = () => {
    const legend = (
      <div className="legend-stack">
        <div className="legend-row">
          <span className="legend-swatch user">
            <span className="legend-icon">👤</span>
          </span>
          <span>User</span>
        </div>
        <div className="legend-row">
          <span className="legend-swatch transaction">
            <span className="legend-icon">💲</span>
          </span>
          <span>Transaction</span>
        </div>
        <div className="legend-row">
          <span className="legend-swatch attribute">
            <span className="legend-icon">🔗</span>
          </span>
          <span>Shared attribute</span>
        </div>
        <div className="legend-row">
          <span className="legend-key direct" />
          <span>Direct link</span>
        </div>
        <div className="legend-row">
          <span className="legend-key linked" />
          <span>Linked transaction</span>
        </div>
      </div>
    );

    if (!selectedNode) {
      return (
        <>
          <h3>Node Details</h3>
          {legend}
          <p className="details-placeholder subtle">Tap a node in the graph to see context.</p>
        </>
      );
    }

    const meta = selectedNode.meta ?? {};

    const rows: Array<{ label: string; value: string }> = [];

    if (selectedNode.type === "user") {
      const userInfo =
        allUsers.find((user) => user.userId === selectedNode.rawId) ?? userSummary;

      rows.push({ label: "User ID", value: selectedNode.rawId });

      if (meta.role) {
        rows.push({ label: "Role", value: meta.role });
      }
      if (meta.direction) {
        rows.push({ label: "Direction", value: meta.direction });
      }
      if (meta.transactionId) {
        rows.push({ label: "Transaction", value: truncateId(meta.transactionId) });
      }
      if (meta.amount !== undefined) {
        rows.push({
          label: "Amount",
          value: formatCurrency(meta.amount, meta.currency)
        });
      }
      if (meta.timestamp) {
        rows.push({ label: "Timestamp", value: formatDate(meta.timestamp) });
      }

      if (meta.nodeKind === "focus" && userRelationships) {
        rows.push({
          label: "Transactions",
          value: `${userRelationships.transactions.length}`
        });
        rows.push({
          label: "Shared Attributes",
          value: `${userRelationships.sharedAttributes.length}`
        });
      }

      if (userInfo) {
        rows.push({ label: "Name", value: userInfo.fullName || "—" });
        rows.push({ label: "Email", value: userInfo.email || "—" });
        rows.push({ label: "Risk", value: describeRiskScore(userInfo.riskScore) });
      }
    } else if (selectedNode.type === "transaction") {
      const txInfo =
        allTransactions.find((tx) => tx.transactionId === selectedNode.rawId) ||
        transactionSummary;

      rows.push({ label: "Transaction ID", value: selectedNode.rawId });

      if (meta.linkType) {
        rows.push({ label: "Link Type", value: meta.linkType });
      }
      if (meta.score !== undefined) {
        rows.push({ label: "Score", value: meta.score.toFixed(2) });
      }
      if (meta.amount !== undefined) {
        rows.push({
          label: "Amount",
          value: formatCurrency(meta.amount, meta.currency)
        });
      }
      if (meta.timestamp) {
        rows.push({ label: "Timestamp", value: formatDate(meta.timestamp) });
      }
      if (meta.updatedAt) {
        rows.push({ label: "Last Updated", value: formatDate(meta.updatedAt) });
      }
      if (txInfo) {
        rows.push({
          label: "Status",
          value: txInfo.status
        });
        rows.push({
          label: "Type",
          value: txInfo.type
        });
      }
    } else if (selectedNode.type === "attribute") {
      rows.push({ label: "Attribute", value: prettifyAttributeType(selectedNode.displayLabel) });
      if (meta.attributeHash) {
        rows.push({ label: "Hash", value: truncateId(meta.attributeHash, 24) });
      }
      if (meta.connectedUsers) {
        rows.push({
          label: "Connected Users",
          value: meta.connectedUsers.length.toString()
        });
      }
    }

    return (
      <>
        <h3>Node Details</h3>
        {legend}
        <div className="details-section">
          <h4 className="details-subheading">Selected Node</h4>
          <ul className="details-list compact">
            {rows.length === 0 ? (
              <li>
                <span className="value">No additional metadata available.</span>
              </li>
            ) : (
              rows.map((row, idx) => (
                <li key={`${row.label}-${idx}`}>
                  <span className="label">{row.label}</span>
                  <span className="value">{row.value}</span>
                </li>
              ))
            )}
          </ul>
        </div>
      </>
    );
  };

  const renderEdgeDetails = () => {
    if (!selectedEdge) {
      return (
        <div className="details-card">
          <h3>Edge Details</h3>
          <p className="details-placeholder subtle">
            Tap an arrow to inspect transaction or attribute links.
          </p>
        </div>
      );
    }

    const meta = selectedEdge.meta ?? {};
    const rows: Array<{ label: string; value: string }> = [];

    const labelMap: Record<string, string> = {
      sourceUserId: "Source user",
      targetUserId: "Target user",
      userId: "User ID",
      transactionId: "Transaction ID",
      sourceTransactionId: "Source transaction",
      targetTransactionId: "Linked transaction",
      attributeType: "Attribute",
      attributeHash: "Attribute hash",
      role: "Role",
      direction: "Direction",
      timestamp: "Timestamp",
      updatedAt: "Last updated",
      linkType: "Link type",
      scope: "Relation type"
    };

    Object.entries(meta).forEach(([key, value]) => {
      if (value === undefined || value === null || value === "") {
        return;
      }
      const label = labelMap[key] || key;
      let display = String(value);
      if (key === "amount") {
        display = formatCurrency(Number(value), (meta as any).currency);
      } else if (key === "timestamp" || key === "updatedAt") {
        display = formatDate(String(value));
      } else if (key === "attributeHash" || key.toLowerCase().includes("id")) {
        display = truncateId(String(value), 22);
      }
      if (key === "currency") {
        return; // handled with amount
      }
      rows.push({ label, value: display });
    });

    return (
      <div className="details-card">
        <h3>Edge Details</h3>
        <p className="details-heading">{selectedEdge.label}</p>
        <ul className="details-list compact">
          {rows.length === 0 ? (
            <li>
              <span className="value">No additional metadata available.</span>
            </li>
          ) : (
            rows.map((row) => (
              <li key={`${row.label}-${row.value}`}>
                <span className="label">{row.label}</span>
                <span className="value">{row.value}</span>
              </li>
            ))
          )}
        </ul>
      </div>
    );
  };

  return (
    <aside className="graph-side">
      <div className="details-card">{renderPrimaryCard()}</div>
      <div className="details-card">{renderNodeDetails()}</div>
      {renderEdgeDetails()}
    </aside>
  );
}
