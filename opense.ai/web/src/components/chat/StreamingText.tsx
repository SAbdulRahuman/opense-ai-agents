"use client";

import { useState, useEffect } from "react";

interface StreamingTextProps {
  text: string;
  speed?: number;
}

export function StreamingText({ text, speed = 15 }: StreamingTextProps) {
  const [displayed, setDisplayed] = useState("");
  const [index, setIndex] = useState(0);

  useEffect(() => {
    if (index < text.length) {
      const timer = setTimeout(() => {
        // Stream in chunks for efficiency
        const chunkSize = Math.min(3, text.length - index);
        setDisplayed(text.slice(0, index + chunkSize));
        setIndex((prev) => prev + chunkSize);
      }, speed);
      return () => clearTimeout(timer);
    }
  }, [index, text, speed]);

  useEffect(() => {
    // Reset when text changes entirely
    setDisplayed("");
    setIndex(0);
  }, [text]);

  return <span>{displayed}</span>;
}
