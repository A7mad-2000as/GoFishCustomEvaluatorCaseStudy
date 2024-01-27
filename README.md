# GoFishCustomEvaluatorCaseStudy

Case study for customizing the evaluation module of the GoFish engine.

For demonstration purposes, the custom evaluation implementaion simply copies the default evalution implementation and changes some bonus scores (in centipawn) as follows:

End game bonus for rook or queen on the seventh rank: 23 --> 45  
Having the bishop pair in the middle game: 22 --> 30  
Having the bishop pair in the end game: 30 --> 45  

Increasing the scores this way encourages the engine to go for an aggressive attacking play style, which could be beneficial or detrimental on the overall playing strength depending on the use case and opponent.
