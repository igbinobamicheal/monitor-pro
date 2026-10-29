export function AmbientBackground() {
  return (
    <div className="fixed inset-0 pointer-events-none z-[-1] overflow-hidden bg-[#ECEAE5]">
      {/* Primary Amber Radiant Glow - Top Center / Right */}
      <div 
        className="absolute -top-[15%] left-[20%] w-[65vw] max-w-[900px] aspect-square rounded-full bg-gradient-to-br from-[#B87C3B]/16 via-[#B87C3B]/6 to-transparent blur-[140px] transform-gpu" 
      />

      {/* Warm Golden Secondary Glow - Center Left */}
      <div 
        className="absolute top-[35%] -left-[15%] w-[55vw] max-w-[750px] aspect-square rounded-full bg-gradient-to-tr from-[#D49852]/12 via-[#B87C3B]/5 to-transparent blur-[150px] transform-gpu" 
      />

      {/* Soft Luminous Bronze Aura - Bottom Right */}
      <div 
        className="absolute -bottom-[20%] right-[5%] w-[60vw] max-w-[800px] aspect-square rounded-full bg-gradient-to-tl from-[#B87C3B]/14 via-[#C98B48]/5 to-transparent blur-[160px] transform-gpu" 
      />

      {/* Subtle Micro Dot Matrix Texture for Tactile Depth */}
      <div 
        className="absolute inset-0 opacity-[0.035] bg-[radial-gradient(#1E1A17_1px,transparent_1px)] [background-size:24px_24px]" 
      />

      {/* Subtle Glass Sheen Vignette */}
      <div className="absolute inset-0 bg-gradient-to-b from-white/20 via-transparent to-black/[0.02]" />
    </div>
  );
}
