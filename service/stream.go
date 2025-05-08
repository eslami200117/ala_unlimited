package service

import (
	"github.com/eslami200117/ala_unlimited/model/extract"
	"github.com/eslami200117/ala_unlimited/model/request"
	pb "github.com/eslami200117/ala_unlimited/protocpb"
	"io"
)

func (c *Core) StreamPrices(stream pb.PriceService_StreamPricesServer) error {
	// Client → Server
	go func() {
		for {
			req, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					close(c.reqQueue)
					return
				}
				c.logger.Error().Err(err).Msg("error receiving from stream")
				return
			}

			converted := request.Request{
				DKP:    req.Dkp,
				Colors: req.Colors,
			}
			c.reqQueue <- converted
		}
	}()

	// Server → Client
	for res := range c.resQueue {

		if err := stream.Send(convertToPb(res)); err != nil {
			c.logger.Error().Err(err).Msg("error sending to stream")
			return err
		}
	}

	return nil
}

func convertToPb(res *extract.ExtProductPrice) *pb.ExtProductPrice {
	converted := &pb.ExtProductPrice{
		Status:      int32(res.Status),
		BuyBoxPrice: int32(res.BuyBoxPrice),
		Variants:    make(map[string]*pb.Variants),
	}

	for k, variantList := range res.Variants {
		protoVariants := &pb.Variants{}
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
