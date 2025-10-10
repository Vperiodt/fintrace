const CURRENCY_REGEX = /^[A-Z]{3}$/;

export const formatCurrency = (amount: number, currency?: string): string => {
  const unit = currency && CURRENCY_REGEX.test(currency) ? currency : "USD";
  try {
    return new Intl.NumberFormat(undefined, {
      style: "currency",
      currency: unit,
      minimumFractionDigits: 2,
      maximumFractionDigits: 2
    }).format(amount);
  } catch {
    return `${amount.toFixed(2)} ${unit}`;
  }
};

export const formatDate = (isoString?: string): string => {
  if (!isoString) {
    return "N/A";
  }
  const date = new Date(isoString);
  if (Number.isNaN(date.getTime())) {
    return isoString;
  }
  return new Intl.DateTimeFormat(undefined, {
    year: "numeric",
    month: "short",
    day: "2-digit"
  }).format(date);
};

export const describeRiskScore = (score?: number): string => {
  if (score === undefined || Number.isNaN(score)) {
    return "N/A";
  }
  const fraction = Math.min(Math.max(score, 0), 1);
  const percentage = Math.round(fraction * 100);

  let tier: string;
  if (fraction < 0.33) {
    tier = "Low";
  } else if (fraction < 0.66) {
    tier = "Medium";
  } else {
    tier = "High";
  }

  return `${tier} (${percentage}%)`;
};

export const truncateId = (value: string, max = 18): string => {
  if (value.length <= max) {
    return value;
  }
  return `${value.slice(0, max - 3)}...`;
};
