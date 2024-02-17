package webserver

import (
	"DemoExchange/internal/app/entities"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Routes struct {
	*Server
	log Logger
}

func (s *Server) NewRoutes() *Routes {
	return &Routes{s, s.log}
}

func (r *Routes) Handler() http.Handler {
	// g := gin.Default()
	g := gin.New()

	g.Use(cors.Default())
	g.Use(gzip.Gzip(gzip.DefaultCompression))
	// g.Use(r.middlewareWhitelistIP())

	g.GET("/stat", r.getStatHandler)

	g.GET("/metrics", func(ctx *gin.Context) {
		promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{
			DisableCompression: true,
		}).ServeHTTP(ctx.Writer, ctx.Request)
	})

	v1 := g.Group("/v1")

	market := v1.Group("/market")
	market.GET("/symbols", r.getMarketSymbolsHandler)
	market.GET("/tickers", r.getMarketTickersHandler)
	market.GET("/orderbook", r.getMarketOrderbookHandler)
	market.GET("/history/orders", r.getMarketHistoryOrdersHandler)

	apikey := v1.Group("/apikey")
	apikey.Use(authSecretMiddleware(r.cfg.AllowServiceTokens))
	apikey.POST("/create", r.postAPIKeyCreateHandler)
	apikey.POST("/disable", r.postAPIKeyDisableHandler)

	wallet := v1.Group("/wallet")
	wallet.Use(r.authTokenMiddleware())
	wallet.GET("/balances", r.getWalletBalancesHandler)
	wallet.POST("/deposit", r.postWalletDepositHandler)
	wallet.POST("/withdraw", r.postWalletWithdrawHandler)

	order := v1.Group("/order")
	order.Use(r.authTokenMiddleware())
	order.POST("/create", r.postOrderCreateHandler)
	order.GET("/get", r.getOrderGetHandler)
	order.POST("/cancel", r.postOrderCancelHandler)
	order.GET("/list", r.getOrderListHandler)

	position := v1.Group("/position")
	position.Use(r.authTokenMiddleware())
	position.GET("/list", r.getPositionListHandler)
	position.POST("/mode", r.postPositionModeHandler)
	position.POST("/type", r.postPositionTypeHandler)
	position.POST("/leverage", r.postPositionLeverageHandler)

	return g
}

func (r *Routes) getStatHandler(c *gin.Context) {
	var stat struct {
		Connections map[uint]int `json:"connections"`
		Messages    int          `json:"messages"`
	}

	// conns := r.srv.service.GetConnections()
	// stat.Connections = make(map[uint]int, len(conns))
	// for _, conn := range conns {
	// 	stat.Connections[conn.UserID]++
	// }

	// stat.Messages, _ = r.srv.service.CountMessages()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"return":  stat,
		"time":    time.Now().Format("2006-01-02 15:04:05"),
	})
}

func (r *Routes) getMarketSymbolsHandler(c *gin.Context) {
	exchange := entities.Exchange(c.Query("exchange"))
	result, err := r.markets.GetMarkets(exchange.Name())
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   err.Error(),
			"time":    time.Now().Format("2006-01-02 15:04:05"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"return":  result,
		"time":    time.Now().Format("2006-01-02 15:04:05"),
	})
}

func (r *Routes) getMarketTickersHandler(c *gin.Context) {
	exchange := entities.Exchange(c.Query("exchange"))
	result, err := r.tickers.GetTickers(exchange.Name())
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   err.Error(),
			"time":    time.Now().Format("2006-01-02 15:04:05"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"return":  result,
		"time":    time.Now().Format("2006-01-02 15:04:05"),
	})
}

func (r *Routes) getMarketOrderbookHandler(c *gin.Context) {
	ctx := c.Request.Context()
	exchange := entities.Exchange(c.Query("exchange"))
	symbol := c.Query("symbol")
	limit := c.DefaultQuery("limit", "100")
	result, err := r.orderbook.GetOrderbook(ctx, exchange.Name(), symbol, limit)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   err.Error(),
			"time":    time.Now().Format("2006-01-02 15:04:05"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"return":  result,
		"time":    time.Now().Format("2006-01-02 15:04:05"),
	})

}

func (r *Routes) getMarketHistoryOrdersHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"return":  make([]entities.Order, 0),
		"time":    time.Now().Format("2006-01-02 15:04:05"),
	})
}

func (r *Routes) postAPIKeyCreateHandler(c *gin.Context) {
	var req CreateTokenRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	ctx := c.Request.Context()
	result, err := r.usecase.CreateToken(ctx, req.Service, req.UserID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   err.Error(),
			"time":    time.Now().Format("2006-01-02 15:04:05"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"return":  result,
		"time":    time.Now().Format("2006-01-02 15:04:05"),
	})
}

func (r *Routes) postAPIKeyDisableHandler(c *gin.Context) {
	var req DisableTokenRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	ctx := c.Request.Context()
	err := r.usecase.DisableToken(ctx, entities.Token(req.Token))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   err.Error(),
			"time":    time.Now().Format("2006-01-02 15:04:05"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"time":    time.Now().Format("2006-01-02 15:04:05"),
	})
}

func (r *Routes) getWalletBalancesHandler(c *gin.Context) {
	ctx := c.Request.Context()
	exchange := c.Query("exchange")
	accountUID, exists := c.Get("accountUID")
	if !exists {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   "Token not found",
			"time":    time.Now().Format("2006-01-02 15:04:05"),
		})
		return
	}

	result, err := r.usecase.GetBalances(ctx, entities.Exchange(exchange), accountUID.(entities.AccountUID))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   err.Error(),
			"time":    time.Now().Format("2006-01-02 15:04:05"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"return":  result,
		"time":    time.Now().Format("2006-01-02 15:04:05"),
	})
}

func (r *Routes) postWalletDepositHandler(c *gin.Context) {
	var req DepositRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	ctx := c.Request.Context()

	accountUID, exists := c.Get("accountUID")
	if !exists {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   "Token not found",
			"time":    time.Now().Format("2006-01-02 15:04:05"),
		})
		return
	}

	amount, err := r.usecase.Deposit(ctx, entities.Exchange(req.Exchange), accountUID.(entities.AccountUID), entities.Coin(req.Coin), req.Amount)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   err.Error(),
			"time":    time.Now().Format("2006-01-02 15:04:05"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"return":  amount,
		"time":    time.Now().Format("2006-01-02 15:04:05"),
	})
}

func (r *Routes) postWalletWithdrawHandler(c *gin.Context) {
	var req WithdrawRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	ctx := c.Request.Context()

	accountUID, exists := c.Get("accountUID")
	if !exists {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   "Token not found",
			"time":    time.Now().Format("2006-01-02 15:04:05"),
		})
		return
	}

	err := r.usecase.Withdraw(ctx, entities.Exchange(req.Exchange), accountUID.(entities.AccountUID), entities.Coin(req.Coin), req.Amount)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   err.Error(),
			"time":    time.Now().Format("2006-01-02 15:04:05"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"time":    time.Now().Format("2006-01-02 15:04:05"),
	})
}

func (r *Routes) postOrderCreateHandler(c *gin.Context) {
	var req OrderCreateRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	ctx := c.Request.Context()

	accountUID, exists := c.Get("accountUID")
	if !exists {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   "Token not found",
			"time":    time.Now().Format("2006-01-02 15:04:05"),
		})
		return
	}

	order := entities.NewOrder(accountUID.(entities.AccountUID))
	order.Exchange = entities.Exchange(req.Exchange)
	order.Symbol = entities.Symbol(req.Symbol)
	order.Type = entities.OrderType(req.Type)
	order.PositionSide = entities.PositionSide(req.PositionSide)
	order.Side = entities.OrderSide(req.Side)
	order.Amount = req.Amount
	order.Price = req.Price

	err := r.usecase.NewOrder(ctx, order)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   err.Error(),
			"time":    time.Now().Format("2006-01-02 15:04:05"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"return":  order,
		"time":    time.Now().Format("2006-01-02 15:04:05"),
	})
}

func (r *Routes) getOrderGetHandler(c *gin.Context) {
	exchange := c.Query("exchange")
	orderUID := c.Query("order_uid")

	ctx := c.Request.Context()

	accountUID, exists := c.Get("accountUID")
	if !exists {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   "Token not found",
			"time":    time.Now().Format("2006-01-02 15:04:05"),
		})
		return
	}

	result, err := r.usecase.GetOrder(ctx, entities.Exchange(exchange), accountUID.(entities.AccountUID), orderUID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   err.Error(),
			"time":    time.Now().Format("2006-01-02 15:04:05"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"return":  result,
		"time":    time.Now().Format("2006-01-02 15:04:05"),
	})
}

func (r *Routes) postOrderCancelHandler(c *gin.Context) {
	var req OrderRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	ctx := c.Request.Context()

	accountUID, exists := c.Get("accountUID")
	if !exists {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   "Token not found",
			"time":    time.Now().Format("2006-01-02 15:04:05"),
		})
		return
	}

	result, err := r.usecase.CancelOrder(ctx, entities.Exchange(req.Exchange), accountUID.(entities.AccountUID), req.OrderUID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   err.Error(),
			"time":    time.Now().Format("2006-01-02 15:04:05"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"return":  result,
		"time":    time.Now().Format("2006-01-02 15:04:05"),
	})
}

func (r *Routes) getOrderListHandler(c *gin.Context) {
	exchange := c.Query("exchange")
	queryStatus := strings.ToLower(c.Query("status"))
	queryLimit := c.DefaultQuery("limit", "100")

	ctx := c.Request.Context()

	accountUID, exists := c.Get("accountUID")
	if !exists {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   "Token not found",
			"time":    time.Now().Format("2006-01-02 15:04:05"),
		})
		return
	}

	limit, err := strconv.Atoi(queryLimit)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   "Limit is wrong value",
			"time":    time.Now().Format("2006-01-02 15:04:05"),
		})
		return
	}

	var statuses []entities.OrderStatus
	switch queryStatus {
	case "open":
		statuses = []entities.OrderStatus{entities.OrderStatusNew, entities.OrderStatusPending}
	case "close":
		statuses = []entities.OrderStatus{entities.OrderStatusSuccess, entities.OrderStatusCancelled, entities.OrderStatusFailed}
	}

	result, err := r.usecase.OrdersList(ctx, entities.Exchange(exchange), accountUID.(entities.AccountUID), statuses, limit)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   err.Error(),
			"time":    time.Now().Format("2006-01-02 15:04:05"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"return":  result,
		"time":    time.Now().Format("2006-01-02 15:04:05"),
	})
}

func (r *Routes) getPositionListHandler(c *gin.Context) {
	accountUID, exists := c.Get("accountUID")
	if !exists {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   "Token not found",
			"time":    time.Now().Format("2006-01-02 15:04:05"),
		})
		return
	}

	exchange := c.Query("exchange")

	ctx := c.Request.Context()

	result, err := r.usecase.PositionsList(ctx, entities.Exchange(exchange), accountUID.(entities.AccountUID))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   err.Error(),
			"time":    time.Now().Format("2006-01-02 15:04:05"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"return":  result,
		"time":    time.Now().Format("2006-01-02 15:04:05"),
	})
}

func (r *Routes) postPositionModeHandler(c *gin.Context) {
	var req PositionModeRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	ctx := c.Request.Context()

	accountUID, exists := c.Get("accountUID")
	if !exists {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   "Token not found",
			"time":    time.Now().Format("2006-01-02 15:04:05"),
		})
		return
	}

	err := r.usecase.SetAccountPositionMode(ctx, entities.Exchange(req.Exchange), accountUID.(entities.AccountUID), entities.PositionMode(req.Mode))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   err.Error(),
			"time":    time.Now().Format("2006-01-02 15:04:05"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"time":    time.Now().Format("2006-01-02 15:04:05"),
	})
}

func (r *Routes) postPositionTypeHandler(c *gin.Context) {
	var req PositionTypeRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	ctx := c.Request.Context()

	accountUID, exists := c.Get("accountUID")
	if !exists {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   "Token not found",
			"time":    time.Now().Format("2006-01-02 15:04:05"),
		})
		return
	}

	err := r.usecase.SetPositionMarginType(ctx, entities.Exchange(req.Exchange), accountUID.(entities.AccountUID), entities.Symbol(req.Symbol), entities.MarginType(req.Type))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   err.Error(),
			"time":    time.Now().Format("2006-01-02 15:04:05"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"time":    time.Now().Format("2006-01-02 15:04:05"),
	})
}

func (r *Routes) postPositionLeverageHandler(c *gin.Context) {
	var req PositionLeverageRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	ctx := c.Request.Context()

	accountUID, exists := c.Get("accountUID")
	if !exists {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   "Token not found",
			"time":    time.Now().Format("2006-01-02 15:04:05"),
		})
		return
	}

	err := r.usecase.SetPositionLeverage(ctx, entities.Exchange(req.Exchange), accountUID.(entities.AccountUID), entities.Symbol(req.Symbol), entities.PositionLeverage(req.Leverage))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   err.Error(),
			"time":    time.Now().Format("2006-01-02 15:04:05"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"time":    time.Now().Format("2006-01-02 15:04:05"),
	})
}
