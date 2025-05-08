package service

import (
	"github.com/eslami200117/ala_unlimited/model/extract"
	"github.com/eslami200117/ala_unlimited/model/request"
	pb "github.com/eslami200117/ala_unlimited/protocpb"
	"io"
	"sync"
)

func (c *Core) StreamPrices(stream pb.PriceService_StreamPricesServer) error {
	ctx := stream.Context()
	var wg sync.WaitGroup
	errCh := make(chan error, 1)

	// Client → Server
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				req, err := stream.Recv()
				if err != nil {
					if err == io.EOF {
						c.logger.Info().Msg("StreamPrices client stream ended")
						close(c.reqQueue)
						return
					}
					c.logger.Error().Err(err).Msg("error receiving from stream")
					errCh <- err
					return
				}

				converted := request.Request{
					DKP:    req.Dkp,
					Colors: req.Colors,
				}

				select {
				case c.reqQueue <- converted:
					// Successfully sent to queue
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	// Server → Client
	go func() {
		wg.Wait()
		// Close errCh after the receiving goroutine is done
		close(errCh)
	}()

	for {
		select {
		case err, ok := <-errCh:
			if !ok {
				// errCh closed, normal exit
				continue
			}
			return err
		case res, ok := <-c.resQueue:
			if !ok {
				// c.resQueue closed, normal exit
				return nil
			}
			if err := stream.Send(convertToPb(res)); err != nil {
				c.logger.Error().Err(err).Msg("error sending to stream")
				return err
			}
		case <-ctx.Done():
			c.logger.Info().Msg("StreamPrices context canceled")
			return ctx.Err()
		}
	}
}

func convertToPb(res *extract.ExtProductPrice) *pb.ExtProductPrice {
	if res == nil {
		return nil
	}

	converted := &pb.ExtProductPrice{
		Status:      int32(res.Status),
		BuyBoxPrice: int32(res.BuyBoxPrice),
		Variants:    make(map[string]*pb.Variants),
	}

	for k, variantList := range res.Variants {
		protoVariants := &pb.Variants{
			Items: make([]*pb.Variant, 0, len(variantList)),
		}
		for _, v := range variantList {
			protoVariants.Items = append(protoVariants.Items, &pb.Variant{
				Seller:         v.Seller,
				SellerId:       int32(v.SellerID),
				Price:          int32(v.Price),
				VarWiner:       v.VarWiner,
				BuyBoxSellerId: int32(v.BuyBoxSellerID),
				Promotion:      v.Promotion,
			})
		}
		converted.Variants[k] = protoVariants
	}
	return converted
}
