interface PaginationControlsProps {
  page: number;
  totalPages: number;
  onPageChange: (page: number) => void;
  disabled?: boolean;
}

export function PaginationControls({
  page,
  totalPages,
  onPageChange,
  disabled
}: PaginationControlsProps) {
  const canPrev = page > 1;
  const canNext = totalPages === 0 ? false : page < totalPages;

  return (
    <div className="pagination-controls">
      <button
        type="button"
        onClick={() => onPageChange(page - 1)}
        disabled={!canPrev || disabled}
      >
        Previous
      </button>
      <span>
        Page {totalPages === 0 ? 0 : page} / {totalPages}
      </span>
      <button
        type="button"
        onClick={() => onPageChange(page + 1)}
        disabled={!canNext || disabled}
      >
        Next
      </button>
    </div>
  );
}
