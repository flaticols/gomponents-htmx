import { BaseElement } from './base-element';

/**
 * A customizable button component.
 * Extends BaseElement to inherit theming capabilities.
 *
 * @customElement my-button
 * @fires click - Fired when the button is clicked
 * @fires focus - Fired when the button receives focus
 * @slot icon - Slot for an optional icon
 * @slot - Default slot for button text
 * @cssProperty --button-padding - Padding inside the button
 * @cssProperty --button-border-radius - Border radius of the button
 * @cssPart button - The button element itself
 */
export class ButtonElement extends BaseElement {
  /**
   * The button variant
   * @attr
   */
  variant: 'primary' | 'secondary' | 'danger' = 'primary';

  /**
   * Button size
   * @attr size
   */
  size: 'small' | 'medium' | 'large' = 'medium';

  /**
   * Whether the button is in loading state
   * @attr
   */
  loading: boolean = false;

  /**
   * Click handler callback
   */
  onClick?: (event: MouseEvent) => void;

  /**
   * Simulates a button click
   * @returns The click event that was dispatched
   */
  click(): MouseEvent {
    const event = new MouseEvent('click');
    this.dispatchEvent(event);
    return event;
  }

  /**
   * Focuses the button element
   */
  override focus(): void {
    super.focus();
    this.dispatchEvent(new FocusEvent('focus'));
  }

  /**
   * Renders the button with the given content
   * @param content - Content to render inside the button
   * @param options - Rendering options
   * @returns The rendered HTML string
   */
  render<T extends object>(content: string, options?: T): string {
    return `<button class="${this.variant} ${this.size}">${content}</button>`;
  }
}

/**
 * Icon button with only an icon, no text
 * @customElement icon-button
 */
export class IconButtonElement extends ButtonElement {
  /**
   * The icon name to display
   * @attr
   */
  icon: string = '';

  /**
   * Accessibility label for the button
   * @attr aria-label
   */
  ariaLabel: string = '';
}
