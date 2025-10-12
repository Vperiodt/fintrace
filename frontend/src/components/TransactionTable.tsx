import { TransactionSummary, Pagination } from "../api/types";
import { PaginationControls } from "./PaginationControls";

interface TransactionTableProps {
  transactions: TransactionSummary[];
  loading: boolean;
  error?: string | null;
  pagination?: Pagination;
  onPageChange: (page: number) => void;
  onSelect: (transactionId: string) => void;
  selectedTransactionId?: string | null;
}

export function TransactionTable({
  transactions,
  loading,
  error,
  pagination,
  onPageChange,
  onSelect,
  selectedTransactionId
}: TransactionTableProps) {
  return (
    <div className="table-wrapper">
      {error && <div className="error">{error}</div>}
      <div className="table-container">
        <table>
          <thead>
            <tr>
              <th>Transaction ID</th>
              <th>Sender</th>
              <th>Receiver</th>
              <th>Amount</th>
              <th>Status</th>
              <th>Type</th>
              <th>Timestamp</th>
            </tr>
          </thead>
          <tbody>
            {loading && (
              <tr>
                <td colSpan={7} className="placeholder-row">
                  Loading transactions...
                </td>
              </tr>
            )}
            {!loading && transactions.length === 0 && (
              <tr>
                <td colSpan={7} className="placeholder-row">
                  No transactions found.
                </td>
              </tr>
            )}
            {!loading &&
              transactions.map((tx) => (
                <tr
                  key={tx.transactionId}
                  className={tx.transactionId === selectedTransactionId ? "selected" : ""}
                  onClick={() => onSelect(tx.transactionId)}
                >
                  <td>{tx.transactionId}</td>
                  <td>{tx.senderUserId}</td>
                  <td>{tx.receiverUserId}</td>
                  <td>
                    {tx.amount.toLocaleString(undefined, {
                      minimumFractionDigits: 2,
                      maximumFractionDigits: 2
                    })} {tx.currency}
                  </td>
                  <td>{tx.status}</td>
                  <td>{tx.type}</td>
                  <td>{new Date(tx.timestamp).toLocaleString()}</td>
                </tr>
              ))}
          </tbody>
        </table>
      </div>
      {pagination && (
        <PaginationControls
          page={pagination.page}
          totalPages={pagination.totalPages}
          onPageChange={onPageChange}
          disabled={loading}
        />
      )}
    </div>
  );
}
