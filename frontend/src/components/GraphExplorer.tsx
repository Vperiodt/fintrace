import { useEffect, useMemo, useRef } from "react";
import cytoscape, { ElementDefinition } from "cytoscape";
import {
  UserRelationshipsResponse,
  TransactionRelationshipsResponse
} from "../api/types";
import { formatCurrency, truncateId } from "../utils/format";

export interface GraphNodeSelection {
  id: string;
  rawId: string;
  displayLabel: string;
  type: "user" | "transaction" | "attribute";
  meta?: Record<string, any>;
}

export interface GraphEdgeSelection {
  id: string;
  label: string;
  meta?: Record<string, any>;
}

interface GraphExplorerProps {
  userRelationships?: UserRelationshipsResponse | null;
  transactionRelationships?: TransactionRelationshipsResponse | null;
  mode: "user" | "transaction" | null;
  loading: boolean;
  error?: string | null;
  onNodeSelect?: (selection: GraphNodeSelection | null) => void;
  onEdgeSelect?: (selection: GraphEdgeSelection | null) => void;
  searchTerm?: string;
}

const GRAPH_FONT = '"Manrope", "Inter", "Segoe UI", sans-serif';

const makeIcon = (background: string, glyph: string) => {
  const svg = `
    <svg xmlns="http://www.w3.org/2000/svg" width="96" height="96" viewBox="0 0 96 96">
      <defs>
        <filter id="shadow" x="-20%" y="-20%" width="140%" height="140%">
          <feDropShadow dx="0" dy="3" stdDeviation="4" flood-color="rgba(15,23,42,0.25)"/>
        </filter>
      </defs>
      <circle cx="48" cy="48" r="40" fill="${background}" filter="url(#shadow)"/>
      <text x="48" y="58" text-anchor="middle" font-family="Manrope, Arial, sans-serif" font-size="38" fill="#ffffff">${glyph}</text>
    </svg>`;
  return `data:image/svg+xml;utf8,${encodeURIComponent(svg)}`;
};

const NODE_ICONS = {
  user: makeIcon("#2563eb", "👤"),
  transaction: makeIcon("#059669", "💲"),
  attribute: makeIcon("#d97706", "🔗")
};

const GRAPH_STYLES: cytoscape.Stylesheet[] = [
  {
    selector: "node",
    style: {
      label: "data(label)",
      color: "#0f172a",
      "font-family": GRAPH_FONT,
      "font-size": 12,
      "font-weight": 600,
      "text-valign": "top",
      "text-halign": "center",
      "text-wrap": "wrap",
      "text-max-width": 140,
      "text-margin-y": -78,
      "text-background-color": "rgba(255,255,255,0.96)",
      "text-background-opacity": 1,
      "text-background-padding": 4,
      "text-border-width": 1,
      "text-border-color": "#cbd5f5",
      "text-border-opacity": 1,
      width: 76,
      height: 76,
      "border-width": 3,
      "border-color": "#e2e8f0",
      "background-image": "data(icon)",
      "background-fit": "cover",
      "transition-property": "border-color border-width background-color",
      "transition-duration": "180ms"
    }
  },
  {
    selector: "node.hover",
    style: {
      "border-color": "#1e293b",
      "border-width": 5,
      "z-index": 999
    }
  },
  {
    selector: "node:selected",
    style: {
      "border-color": "#1e293b",
      "border-width": 6,
      "box-shadow": "0 0 18px rgba(30,41,59,0.35)"
    }
  },
  {
    selector: "node.faded",
    style: {
      opacity: 0.2
    }
  },
  {
    selector: "edge",
    style: {
      width: 4,
      "curve-style": "bezier",
      "target-arrow-shape": "triangle",
      "arrow-scale": 1.2,
      label: "data(label)",
      "font-family": GRAPH_FONT,
      "font-size": 11,
      "line-color": "#64748b",
      "target-arrow-color": "#64748b",
      "text-background-color": "rgba(255,255,255,0.92)",
      "text-background-padding": 3,
      "text-border-width": 1,
      "text-border-color": "#cbd5f5",
      "text-border-opacity": 1,
      "text-wrap": "wrap",
      "text-max-width": 180,
      "text-rotation": "autorotate",
      "text-margin-y": -16
    }
  },
  {
    selector: 'edge[edgeType="direct"]',
    style: {
      "line-color": "#0ea5e9",
      "target-arrow-color": "#0ea5e9",
      width: 4
    }
  },
  {
    selector: 'edge[edgeType="attribute"]',
    style: {
      "line-color": "#f97316",
      "target-arrow-color": "#f97316",
      "line-style": "dashed"
    }
  },
  {
    selector: 'edge[edgeType="linked"]',
    style: {
      "line-color": "#22c55e",
      "target-arrow-color": "#22c55e"
    }
  },
  {
    selector: "edge.hover",
    style: {
      "line-color": "#1f2937",
      "target-arrow-color": "#1f2937",
      "font-weight": 600
    }
  },
  {
    selector: "edge.faded",
    style: {
      opacity: 0.2
    }
  }
];

export function GraphExplorer({
  userRelationships,
  transactionRelationships,
  mode,
  loading,
  error,
  onNodeSelect,
  onEdgeSelect,
  searchTerm
}: GraphExplorerProps) {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const cyRef = useRef<cytoscape.Core | null>(null);
  const nodeSelectionCallback = useRef(onNodeSelect);
  const edgeSelectionCallback = useRef(onEdgeSelect);

  useEffect(() => {
    nodeSelectionCallback.current = onNodeSelect;
  }, [onNodeSelect]);

  useEffect(() => {
    edgeSelectionCallback.current = onEdgeSelect;
  }, [onEdgeSelect]);

  const emitNodeSelection = (node: cytoscape.NodeSingular | null) => {
    const cb = nodeSelectionCallback.current;
    if (!cb) return;
    if (!node) {
      cb(null);
      return;
    }
    const data = node.data();
    cb({
      id: data.id,
      rawId: data.rawId ?? data.rawLabel ?? data.id,
      displayLabel: data.rawLabel ?? data.label ?? data.id,
      type: data.type,
      meta: {
        ...data.meta,
        nodeKind: data.nodeKind
      }
    });
  };

  const emitEdgeSelection = (edge: cytoscape.EdgeSingular | null) => {
    const cb = edgeSelectionCallback.current;
    if (!cb) return;
    if (!edge) {
      cb(null);
      return;
    }
    const data = edge.data();
    cb({
      id: data.id,
      label: data.rawLabel ?? data.label ?? data.id,
      meta: data.meta
    });
  };

  useEffect(() => {
    if (!containerRef.current) {
      return;
    }

    const cy = cytoscape({
      container: containerRef.current,
      style: GRAPH_STYLES,
      boxSelectionEnabled: false,
      selectionType: "single",
      wheelSensitivity: 0.2
    });

    cy.userZoomingEnabled(true);
    cy.userPanningEnabled(true);
    cy.autoungrabify(false);
    cyRef.current = cy;

    const handleNodeHover = (event: cytoscape.EventObject) => {
      event.target.addClass("hover");
    };
    const handleNodeOut = (event: cytoscape.EventObject) => {
      event.target.removeClass("hover");
    };
    const handleEdgeHover = (event: cytoscape.EventObject) => {
      event.target.addClass("hover");
    };
    const handleEdgeOut = (event: cytoscape.EventObject) => {
      event.target.removeClass("hover");
    };

    cy.on("tap", "node", (event) => {
      const node = event.target as cytoscape.NodeSingular;
      cy.nodes().unselect();
      cy.edges().unselect();
      node.select();
      emitNodeSelection(node);
      emitEdgeSelection(null);
    });

    cy.on("tap", "edge", (event) => {
      const edge = event.target as cytoscape.EdgeSingular;
      cy.edges().unselect();
      cy.nodes().unselect();
      edge.select();
      emitEdgeSelection(edge);
      emitNodeSelection(null);
    });

    cy.on("tap", (event) => {
      if (event.target === cy) {
        cy.elements().unselect();
        emitNodeSelection(null);
        emitEdgeSelection(null);
      }
    });

    cy.on("mouseover", "node", handleNodeHover);
    cy.on("mouseout", "node", handleNodeOut);
    cy.on("mouseover", "edge", handleEdgeHover);
    cy.on("mouseout", "edge", handleEdgeOut);

    return () => {
      cy.off("mouseover", "node", handleNodeHover);
      cy.off("mouseout", "node", handleNodeOut);
      cy.off("mouseover", "edge", handleEdgeHover);
      cy.off("mouseout", "edge", handleEdgeOut);
      cy.off("tap");
      cy.destroy();
      cyRef.current = null;
    };
  }, []);

  useEffect(() => {
    const cy = cyRef.current;
    if (!cy) return;

    cy.elements().remove();

    let elements: ElementDefinition[] = [];
    if (mode === "user" && userRelationships) {
      elements = buildUserGraph(userRelationships);
    } else if (mode === "transaction" && transactionRelationships) {
      elements = buildTransactionGraph(transactionRelationships);
    }

    if (elements.length > 0) {
      cy.add(elements);
      cy.layout({ name: "cose", animate: false, nodeRepulsion: 9000 }).run();
      cy.fit(undefined, 50);

      const focusNode = cy.nodes('[nodeKind = "focus"]').first();
      if (focusNode && focusNode.nonempty()) {
        focusNode.select();
        emitNodeSelection(focusNode);
        emitEdgeSelection(null);
      } else {
        const firstNode = cy.nodes().first();
        if (firstNode && firstNode.nonempty()) {
          firstNode.select();
          emitNodeSelection(firstNode);
          emitEdgeSelection(null);
        }
      }
    } else {
      emitNodeSelection(null);
      emitEdgeSelection(null);
    }
  }, [mode, userRelationships, transactionRelationships]);

  useEffect(() => {
    const cy = cyRef.current;
    if (!cy) return;
    const term = searchTerm?.trim().toLowerCase();

    if (!term) {
      cy.elements().removeClass("faded");
      return;
    }

    const matchedNodes = cy.nodes().filter((node) => {
      const raw = (node.data("rawLabel") || node.data("label") || "").toString().toLowerCase();
      return raw.includes(term);
    });

    const matchedEdges = cy.edges().filter((edge) => {
      const raw = (edge.data("rawLabel") || edge.data("label") || "").toString().toLowerCase();
      if (raw.includes(term)) return true;
      const sourceMatch = matchedNodes.some((node) => node.id() === edge.source().id());
      const targetMatch = matchedNodes.some((node) => node.id() === edge.target().id());
      return sourceMatch || targetMatch;
    });

    cy.elements().addClass("faded");
    matchedNodes.removeClass("faded");
    matchedEdges.removeClass("faded");
    matchedEdges.connectedNodes().removeClass("faded");
  }, [searchTerm]);

  const message = useMemo(() => {
    if (loading) {
      return "Loading graph...";
    }
    if (error) {
      return error;
    }
    if (mode === "user" && userRelationships) {
      if (
        userRelationships.directConnections.length === 0 &&
        userRelationships.transactions.length === 0 &&
        userRelationships.sharedAttributes.length === 0
      ) {
        return "No relationships found for the selected user.";
      }
      return "";
    }
    if (mode === "transaction" && transactionRelationships) {
      if (
        transactionRelationships.users.length === 0 &&
        transactionRelationships.linkedTransactions.length === 0
      ) {
        return "No relationships found for the selected transaction.";
      }
      return "";
    }
    return "Select a user or transaction to explore relationships.";
  }, [loading, error, mode, userRelationships, transactionRelationships]);

  return (
    <div className="graph-container">
      <div className="graph-legend">
        <span className="legend-item user">
          <span className="legend-icon">👤</span>User
        </span>
        <span className="legend-item transaction">
          <span className="legend-icon">💲</span>Transaction
        </span>
        <span className="legend-item attribute">
          <span className="legend-icon">🔗</span>Shared Attribute
        </span>
        <span className="legend-item direct">Direct Link</span>
        <span className="legend-item linked">Linked Transaction</span>
      </div>
      <div className="graph-view" ref={containerRef} />
      {message && <div className="graph-placeholder">{message}</div>}
    </div>
  );
}

const buildUserGraph = (data: UserRelationshipsResponse): ElementDefinition[] => {
  const elements: ElementDefinition[] = [];
  const userId = data.userId;
  const focusNodeId = `user:${userId}`;

  elements.push({
    data: {
      id: focusNodeId,
      label: formatNodeLabel(userId),
      rawLabel: userId,
      rawId: userId,
      type: "user",
      nodeKind: "focus",
      icon: NODE_ICONS.user,
      meta: {
        scope: "user",
        transactionCount: data.transactions.length,
        sharedAttributeCount: data.sharedAttributes.length
      }
    }
  });

  const existing = new Set<string>([focusNodeId]);

  data.directConnections.forEach((link, index) => {
    const peerId = `user:${link.userId}`;
    if (!existing.has(peerId)) {
      existing.add(peerId);
      elements.push({
        data: {
          id: peerId,
          label: formatNodeLabel(link.userId),
          rawLabel: link.userId,
          rawId: link.userId,
          type: "user",
          nodeKind: "related",
          icon: NODE_ICONS.user,
          meta: {
            scope: "directConnection",
            linkType: link.linkType,
            direction: link.direction,
            transactionId: link.transactionId,
            amount: link.amount,
            currency: link.currency,
            timestamp: link.timestamp
          }
        }
      });
    }

    elements.push({
      data: {
        id: `direct:${index}`,
        source: focusNodeId,
        target: peerId,
        label: formatEdgeLabel(link.linkType, {
          amount: link.amount,
          currency: link.currency
        }),
        rawLabel: `${link.linkType} · ${formatCurrency(link.amount, link.currency)}`,
        edgeType: "direct",
        meta: {
          scope: "directConnection",
          sourceUserId: userId,
          targetUserId: link.userId,
          transactionId: link.transactionId,
          amount: link.amount,
          currency: link.currency,
          timestamp: link.timestamp
        }
      }
    });
  });

  data.transactions.forEach((tx, index) => {
    const txNodeId = `tx:${tx.transactionId}`;
    if (!existing.has(txNodeId)) {
      existing.add(txNodeId);
      elements.push({
        data: {
          id: txNodeId,
          label: formatNodeLabel(tx.transactionId),
          rawLabel: tx.transactionId,
          rawId: tx.transactionId,
          type: "transaction",
          nodeKind: "related",
          icon: NODE_ICONS.transaction,
          meta: {
            scope: "userTransaction",
            role: tx.role,
            amount: tx.amount,
            currency: tx.currency,
            timestamp: tx.timestamp
          }
        }
      });
    }

    elements.push({
      data: {
        id: `participation:${index}`,
        source: focusNodeId,
        target: txNodeId,
        label: formatEdgeLabel(tx.role, {
          amount: tx.amount,
          currency: tx.currency
        }),
        rawLabel: `${tx.role} · ${formatCurrency(tx.amount, tx.currency)}`,
        edgeType: "direct",
        meta: {
          scope: "userTransaction",
          userId: userId,
          transactionId: tx.transactionId,
          role: tx.role,
          amount: tx.amount,
          currency: tx.currency,
          timestamp: tx.timestamp
        }
      }
    });
  });

  data.sharedAttributes.forEach((attr, index) => {
    const attrNodeId = `attr:${attr.attributeType}:${attr.attributeHash}`;
    if (!existing.has(attrNodeId)) {
      existing.add(attrNodeId);
      elements.push({
        data: {
          id: attrNodeId,
          label: formatAttributeLabel(attr.attributeType),
          rawLabel: attr.attributeType,
          rawId: attr.attributeHash,
          type: "attribute",
          nodeKind: "related",
          icon: NODE_ICONS.attribute,
          meta: {
            scope: "sharedAttribute",
            attributeType: attr.attributeType,
            attributeHash: attr.attributeHash,
            connectedUsers: attr.connectedUsers
          }
        }
      });
    }

    elements.push({
      data: {
        id: `attr-link:${index}`,
        source: focusNodeId,
        target: attrNodeId,
        label: formatEdgeLabel("shared"),
        rawLabel: `${attr.attributeType} shared`,
        edgeType: "attribute",
        meta: {
          scope: "sharedAttribute",
          userId: userId,
          attributeType: attr.attributeType,
          attributeHash: attr.attributeHash
        }
      }
    });

    attr.connectedUsers.forEach((connectedUserId, subIdx) => {
      const connectedNodeId = `user:${connectedUserId}`;
      if (!existing.has(connectedNodeId)) {
        existing.add(connectedNodeId);
        elements.push({
          data: {
            id: connectedNodeId,
            label: formatNodeLabel(connectedUserId),
            rawLabel: connectedUserId,
            rawId: connectedUserId,
            type: "user",
            nodeKind: "related",
            icon: NODE_ICONS.user,
            meta: {
              scope: "sharedAttributePeer",
              attributeType: attr.attributeType,
              attributeHash: attr.attributeHash
            }
          }
        });
      }

      elements.push({
        data: {
          id: `attr-connection:${index}:${subIdx}`,
          source: attrNodeId,
          target: connectedNodeId,
          label: formatAttributeLabel(attr.attributeType),
          rawLabel: `${attr.attributeType} link`,
          edgeType: "attribute",
          meta: {
            scope: "sharedAttributeConnection",
            attributeType: attr.attributeType,
            attributeHash: attr.attributeHash,
            targetUserId: connectedUserId
          }
        }
      });
    });
  });

  return elements;
};

const buildTransactionGraph = (
  data: TransactionRelationshipsResponse
): ElementDefinition[] => {
  const elements: ElementDefinition[] = [];
  const txNodeId = `tx:${data.transactionId}`;
  const existing = new Set<string>();

  elements.push({
    data: {
      id: txNodeId,
      label: formatNodeLabel(data.transactionId),
      rawLabel: data.transactionId,
      rawId: data.transactionId,
      type: "transaction",
      nodeKind: "focus",
      icon: NODE_ICONS.transaction,
      meta: {
        scope: "transaction",
        participantCount: data.users.length,
        linkedTransactionCount: data.linkedTransactions.length
      }
    }
  });
  existing.add(txNodeId);

  data.users.forEach((user, index) => {
    const userNodeId = `user:${user.userId}`;
    if (!existing.has(userNodeId)) {
      existing.add(userNodeId);
      elements.push({
        data: {
          id: userNodeId,
          label: formatNodeLabel(user.userId, user.role),
          rawLabel: user.userId,
          rawId: user.userId,
          type: "user",
          nodeKind: "related",
          icon: NODE_ICONS.user,
          meta: {
            scope: "transactionParticipant",
            role: user.role,
            direction: user.direction,
            amount: user.amount,
            currency: user.currency
          }
        }
      });
    }

    elements.push({
      data: {
        id: `participant:${index}`,
        source: userNodeId,
        target: txNodeId,
        label: formatEdgeLabel(user.direction, {
          amount: user.amount,
          currency: user.currency
        }),
        rawLabel: `${user.direction} · ${formatCurrency(user.amount, user.currency)}`,
        edgeType: "direct",
        meta: {
          scope: "transactionParticipant",
          userId: user.userId,
          transactionId: data.transactionId,
          role: user.role,
          direction: user.direction,
          amount: user.amount,
          currency: user.currency
        }
      }
    });
  });

  data.linkedTransactions.forEach((link, index) => {
    const linkedNodeId = `tx:${link.transactionId}`;
    if (!existing.has(linkedNodeId)) {
      existing.add(linkedNodeId);
      elements.push({
        data: {
          id: linkedNodeId,
          label: formatNodeLabel(link.transactionId),
          rawLabel: link.transactionId,
          rawId: link.transactionId,
          type: "transaction",
          nodeKind: "related",
          icon: NODE_ICONS.transaction,
          meta: {
            scope: "linkedTransaction",
            linkType: link.linkType,
            score: link.score,
            attributeHash: link.attributeHash,
            updatedAt: link.updatedAt
          }
        }
      });
    }

    elements.push({
      data: {
        id: `linked:${index}`,
        source: txNodeId,
        target: linkedNodeId,
        label: formatEdgeLabel(link.linkType, {
          secondary: `score ${link.score.toFixed(2)}`
        }),
        rawLabel: `${link.linkType} · score ${link.score.toFixed(2)}`,
        edgeType: "linked",
        meta: {
          scope: "linkedTransaction",
          linkType: link.linkType,
          attributeHash: link.attributeHash,
          score: link.score,
          updatedAt: link.updatedAt,
          sourceTransactionId: data.transactionId,
          targetTransactionId: link.transactionId
        }
      }
    });
  });

  return elements;
};

const formatNodeLabel = (value: string, secondary?: string): string => {
  const primary = truncateId(value.trim(), 24);
  if (secondary) {
    return `${primary}\n${secondary}`;
  }
  return primary;
};

const formatAttributeLabel = (value: string): string =>
  value
    .toLowerCase()
    .split("_")
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ");

const formatEdgeLabel = (
  primary: string,
  options?: { amount?: number; currency?: string; secondary?: string }
): string => {
  const lines = [primary.replace(/_/g, " ")];
  if (options?.amount !== undefined) {
    lines.push(formatCurrency(options.amount, options.currency));
  }
  if (options?.secondary) {
    lines.push(options.secondary);
  }
  return lines.join("\n");
};
