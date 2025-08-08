package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-playground/validator"
	"github.com/zhavkk/order-service/internal/dto"
	"github.com/zhavkk/order-service/internal/logger"
	"github.com/zhavkk/order-service/internal/repository/postgres"
)

var validate = validator.New()

type OrderService interface {
	GetByID(ctx context.Context, req *dto.GetOrderByIDRequest) (*dto.GetOrderByIDResponse, error)
	ProcessMessage(ctx context.Context, message []byte) error
	ProcessOrder(ctx context.Context, req *dto.ProcessOrderRequest) error
}

type Handler struct {
	orderService OrderService
}

func NewHandler(orderService OrderService) *Handler {
	return &Handler{
		orderService: orderService,
	}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/orders", func(r chi.Router) {
		r.Get("/{order_id}", h.GetOrderByID)
	})
}

func (h *Handler) GetOrderByID(
	w http.ResponseWriter,
	r *http.Request,
) {
	const op = "Handler.GetOrderByID"

	var req dto.GetOrderByIDRequest

	req.OrderID = chi.URLParam(r, "order_id")

	if err := validate.Struct(&req); err != nil {
		logger.Log.Error("GetOrderByID", "Invalid request", err)
		h.writeErrorResponse(w, "Invalid request", http.StatusBadRequest)
		return
	}

	resp, err := h.orderService.GetByID(r.Context(), &req)
	if err != nil {
		logger.Log.Error(op, "Failed to get order by ID", err)
		if errors.Is(err, postgres.ErrOrderNotFound) {
			h.writeErrorResponse(w, "Order not found", http.StatusNotFound)
			return
		}
		h.writeErrorResponse(w, "Failed to get order", http.StatusInternalServerError)
		return
	}
	h.writeJSONResponse(w, resp, http.StatusOK)

}

func (h *Handler) writeJSONResponse(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.Log.Error("Failed to write JSON response", "error", err)
	}
}

func (h *Handler) writeErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	errorResp := dto.ErrorResponse{
		Error:   message,
		Message: message,
		Code:    statusCode,
	}
	h.writeJSONResponse(w, errorResp, statusCode)
}
