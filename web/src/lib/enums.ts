export enum BigBoxTypes{
    _,
    Big_Box,
    Small_Box,
    Eidos_Trapezoid,
    DVD_Case_Slipcover,
    Old_Small_Box,
    Box_in_Box,
    Big_Box_With_Gatefold,
    Small_Box_With_Gatefold,
    Big_Box_With_Vertical_Gatefold,
    Big_Box_With_Back_Gatefold,
    New_Small_Box,
    New_Big_Box,
    Small_Box_For_DVD,
    Big_Long_Box,
    Big_Box_With_Vertical_Gatefold_But_Horizontal,
    Small_Box_With_Gatefold_With_Right_Flap,
    DVD_Case_Slipcover_With_Gatefold,
    New_Box_in_box,
    Vinyl_Like_With_Gatefold,
    Vinyl_Like_With_Double_Gatefold,
    Big_Box_With_Front_And_Back_Gatefold,
}

export const NewBigBoxTypes = new Set([BigBoxTypes.New_Small_Box, BigBoxTypes.New_Big_Box,BigBoxTypes.New_Box_in_box])

export const AllGatefoldTypes = new Set([BigBoxTypes.Big_Box_With_Gatefold, BigBoxTypes.Big_Box_With_Vertical_Gatefold, BigBoxTypes.Small_Box_With_Gatefold,BigBoxTypes.Eidos_Trapezoid,BigBoxTypes.Big_Box_With_Back_Gatefold,BigBoxTypes.Big_Box_With_Vertical_Gatefold_But_Horizontal,BigBoxTypes.Small_Box_With_Gatefold_With_Right_Flap, BigBoxTypes.DVD_Case_Slipcover_With_Gatefold,BigBoxTypes.Vinyl_Like_With_Gatefold,BigBoxTypes.Vinyl_Like_With_Double_Gatefold,BigBoxTypes.Big_Box_With_Front_And_Back_Gatefold])
export const NormalGatefoldTypes = new Set([BigBoxTypes.Big_Box_With_Gatefold, BigBoxTypes.Big_Box_With_Vertical_Gatefold, BigBoxTypes.Small_Box_With_Gatefold,BigBoxTypes.Eidos_Trapezoid, BigBoxTypes.DVD_Case_Slipcover_With_Gatefold,BigBoxTypes.Vinyl_Like_With_Gatefold,BigBoxTypes.Vinyl_Like_With_Double_Gatefold,BigBoxTypes.Big_Box_With_Front_And_Back_Gatefold])
export const VerticalGatefoldTypes = new Set([BigBoxTypes.Big_Box_With_Vertical_Gatefold, BigBoxTypes.Eidos_Trapezoid,BigBoxTypes.Big_Box_With_Vertical_Gatefold_But_Horizontal])

export enum ImageOrder{
    front,
    left,
    right,
    back,
    top,
    bottom,
    gatefold0,
    gatefold1,
    gatefold2,
    gatefold3
}

export enum BoxShelfDirection{
    left,
    front
}

export enum GatefoldTypes{
    Not_Gatefold,
    Regular_Gatefold,
    Right_Flap_Gatefold
}