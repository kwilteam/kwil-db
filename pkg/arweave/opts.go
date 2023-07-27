package arweave

import "github.com/everFinance/goar/types"

type BundlrOpts func(*BundlrClient)

func WithTags(tags ...Tag) BundlrOpts {
	return func(b *BundlrClient) {
		goarTags := make([]types.Tag, len(tags))
		for i, tag := range tags {
			goarTags[i] = types.Tag{
				Name:  tag.Name,
				Value: tag.Value,
			}
		}

		b.tags = goarTags
	}
}
